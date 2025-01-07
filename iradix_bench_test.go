package iradix

import (
	"testing"

	"github.com/google/uuid"
)

func benchInsertTxn(b *testing.B, batchSize int, track bool) {
	r := New()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		txn := r.Txn()

		txn.TrackMutate(track)

		keys := make([][]byte, batchSize)
		values := make([]interface{}, batchSize)
		for j := 0; j < batchSize; j++ {
			keys[j] = []byte(uuid.New().String())
			values[j] = j
		}
		txn.BulkInsert(keys, values)

		r = txn.Commit()
	}
}

func Benchmark10InsertTxnTrack(b *testing.B) {
	benchInsertTxn(b, 10, true)
}
func Benchmark10InsertTxnNoTrack(b *testing.B) {
	benchInsertTxn(b, 10, false)
}

func Benchmark100InsertTxnTrack(b *testing.B) {
	benchInsertTxn(b, 100, true)
}
func Benchmark100InsertTxnNoTrack(b *testing.B) {
	benchInsertTxn(b, 100, false)
}

func Benchmark1000InsertTxnTrack(b *testing.B) {
	benchInsertTxn(b, 1000, true)
}
func Benchmark1000InsertTxnNoTrack(b *testing.B) {
	benchInsertTxn(b, 1000, false)
}

func Benchmark10000InsertTxnTrack(b *testing.B) {
	benchInsertTxn(b, 10000, true)
}
func Benchmark10000InsertTxnNoTrack(b *testing.B) {
	benchInsertTxn(b, 10000, false)
}

func Benchmark100000InsertTxnTrack(b *testing.B) {
	benchInsertTxn(b, 100000, true)
}
func Benchmark100000InsertTxnNoTrack(b *testing.B) {
	benchInsertTxn(b, 100000, false)
}
