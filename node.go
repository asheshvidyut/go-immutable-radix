package iradix

import (
	"bytes"
	"math/bits"
)

// WalkFn is used when walking the tree. Takes a
// key and value, returning if iteration should
// be terminated.
type WalkFn func(k []byte, v interface{}) bool

// leafNode is used to represent a value
type leafNode struct {
	mutateCh chan struct{}
	key      []byte
	val      interface{}
}

// edge is used to represent an edge node
type edge struct {
	label byte
	node  *Node
}

// Node is an immutable node in the radix tree
type Node struct {
	// mutateCh is closed if this node is modified
	mutateCh chan struct{}

	// leaf is used to store possible leaf
	leaf *leafNode

	// prefix is the common prefix we ignore
	prefix []byte

	// Edges should be stored in-order for iteration.
	// We avoid a fully materialized slice to save memory,
	// since in most cases we expect to be sparse
	bitmap   [4]uint64
	children []*Node
}

func (n *Node) isLeaf() bool {
	return n.leaf != nil
}

// setBit sets the bit for a given label
func setBit(bitmap *[4]uint64, label byte) {
	block := label / 64
	bitPos := label % 64
	bitmap[block] |= 1 << bitPos
}

// clearBit clears the bit for a given label
func clearBit(bitmap *[4]uint64, label byte) {
	block := label / 64
	bitPos := label % 64
	mask := uint64(1) << bitPos
	bitmap[block] &^= mask
}

// bitSet checks if bit for label is set
func bitSet(bitmap [4]uint64, label byte) bool {
	block := label / 64
	bitPos := label % 64
	return (bitmap[block] & (1 << bitPos)) != 0
}

// rankOf computes how many bits are set before foundLabel
func (n *Node) rankOf(foundLabel uint8) int {
	block := foundLabel / 64
	bitPos := foundLabel % 64
	mask := uint64(1) << bitPos

	rank := 0
	for i := 0; i < int(block); i++ {
		rank += bits.OnesCount64(n.bitmap[i])
	}
	rank += bits.OnesCount64(n.bitmap[block] & (mask - 1))
	return rank
}

// findInsertionIndex finds the index where a label should be inserted.
// Similar to lower bound search in a sorted array, but using a bitmap.
func (n *Node) findInsertionIndex(label byte) int {
	block := label / 64
	bitPos := label % 64

	// Check current block from bitPos upwards
	curBlock := n.bitmap[block] >> bitPos
	if curBlock != 0 {
		// There is at least one set bit >= bitPos in this block
		offset := bits.TrailingZeros64(curBlock)
		foundLabel := uint8(block*64 + bitPos + uint8(offset))
		if foundLabel >= label {
			return n.rankOf(foundLabel)
		}
	}

	// Check subsequent blocks
	for b := block + 1; b < 4; b++ {
		if n.bitmap[b] != 0 {
			offset := bits.TrailingZeros64(n.bitmap[b])
			foundLabel := uint8(b*64 + uint8(offset))
			// foundLabel > label by definition
			return n.rankOf(foundLabel)
		}
	}

	// No existing child >= label, so insert at end
	return len(n.children)
}

func (n *Node) addEdge(label byte, child *Node) {
	idx := n.findInsertionIndex(label)
	n.children = append(n.children, child)
	if idx != len(n.children)-1 {
		copy(n.children[idx+1:], n.children[idx:len(n.children)-1])
		n.children[idx] = child
	}
	setBit(&n.bitmap, label)
}

func (n *Node) getLowerBoundEdge(label byte) (int, *Node) {
	// Similar logic to find the first child with label >= input
	block := label / 64
	bitPos := label % 64

	curBlock := n.bitmap[block] >> bitPos
	if curBlock != 0 {
		offset := bits.TrailingZeros64(curBlock)
		foundLabel := uint8(block*64 + bitPos + uint8(offset))
		rank := n.rankOf(foundLabel)
		return rank, n.children[rank]
	}

	for b := block + 1; b < 4; b++ {
		if n.bitmap[b] != 0 {
			offset := bits.TrailingZeros64(n.bitmap[b])
			foundLabel := uint8(b*64 + uint8(offset))
			rank := n.rankOf(foundLabel)
			return rank, n.children[rank]
		}
	}

	// No child >= label
	return -1, nil
}

func (n *Node) getChildRank(label byte) int {
	block := label / 64
	bitPos := label % 64
	mask := uint64(1) << bitPos

	rank := 0
	for i := 0; i < int(block); i++ {
		rank += bits.OnesCount64(n.bitmap[i])
	}
	rank += bits.OnesCount64(n.bitmap[block] & (mask - 1))
	return rank
}

func (n *Node) replaceEdge(label byte, child *Node) {
	if !bitSet(n.bitmap, label) {
		panic("replacing missing edge")
	}

	// Compute rank
	rank := n.getChildRank(label)
	n.children[rank] = child
}

func (n *Node) getEdge(label byte) (int, *Node) {
	if !bitSet(n.bitmap, label) {
		return -1, nil
	}
	rank := n.getChildRank(label)
	return rank, n.children[rank]
}

func (n *Node) delEdge(label byte) {
	if !bitSet(n.bitmap, label) {
		return
	}
	rank := n.getChildRank(label)
	copy(n.children[rank:], n.children[rank+1:])
	n.children[len(n.children)-1] = nil
	n.children = n.children[:len(n.children)-1]
	clearBit(&n.bitmap, label)
}

func (n *Node) GetWatch(k []byte) (<-chan struct{}, interface{}, bool) {
	search := k
	watch := n.mutateCh
	for {
		// Check for key exhaustion
		if len(search) == 0 {
			if n.isLeaf() {
				return n.leaf.mutateCh, n.leaf.val, true
			}
			break
		}

		// Look for an edge
		_, n = n.getEdge(search[0])
		if n == nil {
			break
		}

		// Update to the finest granularity as the search makes progress
		watch = n.mutateCh

		// Consume the search prefix
		if bytes.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else {
			break
		}
	}
	return watch, nil, false
}

func (n *Node) Get(k []byte) (interface{}, bool) {
	_, val, ok := n.GetWatch(k)
	return val, ok
}

// LongestPrefix is like Get, but instead of an
// exact match, it will return the longest prefix match.
func (n *Node) LongestPrefix(k []byte) ([]byte, interface{}, bool) {
	var last *leafNode
	search := k
	for {
		// Look for a leaf node
		if n.isLeaf() {
			last = n.leaf
		}

		// Check for key exhaution
		if len(search) == 0 {
			break
		}

		// Look for an edge
		_, n = n.getEdge(search[0])
		if n == nil {
			break
		}

		// Consume the search prefix
		if bytes.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else {
			break
		}
	}
	if last != nil {
		return last.key, last.val, true
	}
	return nil, nil, false
}

// Minimum is used to return the minimum value in the tree
func (n *Node) Minimum() ([]byte, interface{}, bool) {
	for {
		if n.isLeaf() {
			return n.leaf.key, n.leaf.val, true
		}
		if len(n.children) > 0 {
			n = n.children[0]
		} else {
			break
		}
	}
	return nil, nil, false
}

// Maximum is used to return the maximum value in the tree
func (n *Node) Maximum() ([]byte, interface{}, bool) {
	for {
		if num := len(n.children); num > 0 {
			n = n.children[num-1]
			continue
		}
		if n.isLeaf() {
			return n.leaf.key, n.leaf.val, true
		} else {
			break
		}
	}
	return nil, nil, false
}

// Iterator is used to return an iterator at
// the given node to walk the tree
func (n *Node) Iterator() *Iterator {
	return &Iterator{node: n}
}

// ReverseIterator is used to return an iterator at
// the given node to walk the tree backwards
func (n *Node) ReverseIterator() *ReverseIterator {
	return NewReverseIterator(n)
}

// rawIterator is used to return a raw iterator at the given node to walk the
// tree.
func (n *Node) rawIterator() *rawIterator {
	iter := &rawIterator{node: n}
	iter.Next()
	return iter
}

// Walk is used to walk the tree
func (n *Node) Walk(fn WalkFn) {
	recursiveWalk(n, fn)
}

// WalkBackwards is used to walk the tree in reverse order
func (n *Node) WalkBackwards(fn WalkFn) {
	reverseRecursiveWalk(n, fn)
}

// WalkPrefix is used to walk the tree under a prefix
func (n *Node) WalkPrefix(prefix []byte, fn WalkFn) {
	search := prefix
	for {
		// Check for key exhaution
		if len(search) == 0 {
			recursiveWalk(n, fn)
			return
		}

		// Look for an edge
		_, n = n.getEdge(search[0])
		if n == nil {
			break
		}

		// Consume the search prefix
		if bytes.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]

		} else if bytes.HasPrefix(n.prefix, search) {
			// Child may be under our search prefix
			recursiveWalk(n, fn)
			return
		} else {
			break
		}
	}
}

// WalkPath is used to walk the tree, but only visiting nodes
// from the root down to a given leaf. Where WalkPrefix walks
// all the entries *under* the given prefix, this walks the
// entries *above* the given prefix.
func (n *Node) WalkPath(path []byte, fn WalkFn) {
	search := path
	for {
		// Visit the leaf values if any
		if n.leaf != nil && fn(n.leaf.key, n.leaf.val) {
			return
		}

		// Check for key exhaution
		if len(search) == 0 {
			return
		}

		// Look for an edge
		_, n = n.getEdge(search[0])
		if n == nil {
			return
		}

		// Consume the search prefix
		if bytes.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else {
			break
		}
	}
}

// recursiveWalk is used to do a pre-order walk of a node
// recursively. Returns true if the walk should be aborted
func recursiveWalk(n *Node, fn WalkFn) bool {
	// Visit the leaf values if any
	if n.leaf != nil && fn(n.leaf.key, n.leaf.val) {
		return true
	}

	// Iterate over children
	for _, child := range n.children {
		if recursiveWalk(child, fn) {
			return true
		}
	}
	return false
}

// reverseRecursiveWalk is used to do a reverse pre-order
// walk of a node recursively. Returns true if the walk
// should be aborted
func reverseRecursiveWalk(n *Node, fn WalkFn) bool {
	// Visit the leaf values if any
	if n.leaf != nil && fn(n.leaf.key, n.leaf.val) {
		return true
	}

	for i := len(n.children) - 1; i >= 0; i-- {
		if reverseRecursiveWalk(n.children[i], fn) {
			return true
		}
	}
	return false
}
