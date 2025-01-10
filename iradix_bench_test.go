package iradix

import (
	"testing"

	"github.com/google/uuid"
)

func benchBulkInsertTxn(b *testing.B, batchSize int, track bool) {
	r := New()

	keys := make([][]byte, b.N*batchSize)
	values := make([]interface{}, b.N*batchSize)

	b.ResetTimer()

	txn := r.Txn()
	txn.TrackMutate(track)

	index := 0
	for i := 0; i < b.N; i++ {
		// Pre-generate UUIDs to avoid runtime overhead
		for j := 0; j < batchSize; j++ {
			keys[index] = []byte(uuid.New().String())
			values[index] = j
			index++
		}
	}
	txn.BulkInsert(keys, values)

	r = txn.Commit()
}

func Benchmark10BulkInsertTxnTrack(b *testing.B) {
	benchBulkInsertTxn(b, 10, true)
}
func Benchmark10BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 10, false)
}

func Benchmark100BulkInsertTxnTrack(b *testing.B) {
	benchBulkInsertTxn(b, 100, true)
}
func Benchmark100BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 100, false)
}

func Benchmark1000BulkInsertTxnTrack(b *testing.B) {
	benchBulkInsertTxn(b, 1000, true)
}
func Benchmark1000BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 1000, false)
}

func Benchmark10000BulkInsertTxnTrack(b *testing.B) {
	benchBulkInsertTxn(b, 10000, true)
}
func Benchmark10000BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 10000, false)
}

func Benchmark100000BulkInsertTxnTrack(b *testing.B) {
	benchBulkInsertTxn(b, 100000, true)
}
func Benchmark100000BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 100000, false)
}

func Benchmark1000000BulkInsertTxnTrack(b *testing.B) {
	benchBulkInsertTxn(b, 1000000, true)
}
func Benchmark1000000BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 1000000, false)
}
