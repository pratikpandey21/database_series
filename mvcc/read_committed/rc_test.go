package read_committed

import (
	"testing"
)

func TestReadCommitted(t *testing.T) {
	store := NewMVCCStore()

	key := "testKey"
	initialValue := "initial"

	transactionInitial := store.Begin()
	if !store.Write(key, initialValue, transactionInitial) {
		t.Fatal("Initial Transaction failed to write")
	}

	if !store.Commit(transactionInitial) {
		t.Fatal("Initial Transaction failed to commit")
	}

	transactionA := store.Begin()
	newValue := "updated"
	if !store.Write(key, newValue, transactionA) {
		t.Fatal("Transaction A failed to write")
	}

	transactionB := store.Begin()
	if !store.Commit(transactionA) {
		t.Fatal("Transaction A failed to commit")
	}

	if value, ok := store.Read(key); !ok || value != newValue {
		t.Fatalf("Transaction B read unexpected value after A commit: got %v, want %v", value, newValue)
	}

	if !store.Rollback(transactionB) {
		t.Fatal("Transaction B failed to rollback")
	}
}
