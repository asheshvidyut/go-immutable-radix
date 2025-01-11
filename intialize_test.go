package iradix

import (
	"bufio"
	"log"
	"os"
	"testing"
)

func TestInitializeWithData(t *testing.T) {
	file, err := os.Open("words.txt")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer file.Close()
	keys := make([][]byte, 0)
	values := make([]interface{}, 0)
	// Create a scanner
	scanner := bufio.NewScanner(file)
	index := 0
	for scanner.Scan() {
		// Print each line
		keys = append(keys, []byte(scanner.Text()))
		values = append(values, index)
		index++
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	r := NewWithData(keys, values)

	if r.Len() != len(keys) {
		t.Fatalf("expected %d, got %d", len(keys), r.Len())
	}
	for idx, key := range keys {
		if val, ok := r.Get(key); !ok || val != values[idx] {
			t.Fatalf("expected %v, got %v", values[idx], val)
		}
	}
}
