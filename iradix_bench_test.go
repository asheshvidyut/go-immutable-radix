package iradix

import (
	"testing"

	"github.com/google/uuid"
)

func benchInsertTxn(b *testing.B, batchSize int, track bool) {
	r := New()

	// Pre-generate keys
	keys := make([][]byte, b.N*batchSize)
	for i := 0; i < b.N*batchSize; i++ {
		keys[i] = []byte(uuid.New().String())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		txn := r.Txn()
		txn.TrackMutate(track)

		// Insert keys for this batch
		start := i * batchSize
		for j := 0; j < batchSize; j++ {
			txn.Insert(keys[start+j], j)
		}

		r = txn.Commit()
	}
}

func benchBulkInsertTxn(b *testing.B, batchSize int, track bool) {
	r := New()

	// Pre-generate keys and values
	keys := make([][]byte, b.N*batchSize)
	values := make([]interface{}, b.N*batchSize)
	for i := 0; i < b.N*batchSize; i++ {
		keys[i] = []byte(uuid.New().String())
		values[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := i * batchSize
		txn := r.Txn()
		txn.InitializeWithData(keys[start:start+batchSize], values[start:start+batchSize])
		r = txn.Commit()
	}

}

func Benchmark10InsertTxnNoTrack(b *testing.B) {
	benchInsertTxn(b, 10, false)
}

func Benchmark10BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 10, false)
}

func Benchmark100InsertTxnNoTrack(b *testing.B) {
	benchInsertTxn(b, 100, false)
}

func Benchmark100BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 100, false)
}

func Benchmark1000InsertTxnNoTrack(b *testing.B) {
	benchInsertTxn(b, 1000, false)
}

func Benchmark1000BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 1000, false)
}

func Benchmark10000InsertTxnNoTrack(b *testing.B) {
	benchInsertTxn(b, 10000, false)
}

func Benchmark10000BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 10000, false)
}

func Benchmark100000InsertTxnNoTrack(b *testing.B) {
	benchInsertTxn(b, 100000, false)
}

func Benchmark100000BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 100000, false)
}

func Benchmark1000000InsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 1000000, false)
}

func Benchmark1000000BulkInsertTxnNoTrack(b *testing.B) {
	benchInsertTxn(b, 1000000, false)
}
