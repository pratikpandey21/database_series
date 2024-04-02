package repeatable_read

import (
	"math/rand"
	"sync"
	"testing"
)

func TestTransactionCommit(t *testing.T) {
	store := NewMVCCStore()
	transactionID := store.Begin()
	key := "testKey"
	value := "testValue"

	store.Write(key, value, transactionID)

	if !store.Commit(transactionID) {
		t.Fatalf("Failed to commit transaction")
	}

	readValue, found := store.Read(key, transactionID)
	if !found || readValue != value {
		t.Errorf("Expected to find value '%s', but got '%v'", value, readValue)
	}
}

func TestTransactionRollback(t *testing.T) {
	store := NewMVCCStore()
	transactionID := store.Begin()
	key := "testKey"
	value := "testValue"

	store.Write(key, value, transactionID)
	store.Rollback(transactionID)

	_, found := store.Read(key, transactionID)
	if found {
		t.Error("Expected not to find value after rollback")
	}
}

func TestTransactionIsolation(t *testing.T) {
	store := NewMVCCStore()
	txn1 := store.Begin()
	txn2 := store.Begin()

	key := "testKey"
	valueTxn1 := "valueTxn1"
	valueTxn2 := "valueTxn2"

	store.Write(key, valueTxn1, txn1)
	store.Write(key, valueTxn2, txn2)

	readValueTxn1, foundTxn1 := store.Read(key, txn1)
	if !foundTxn1 || readValueTxn1 != valueTxn1 {
		t.Errorf("Txn1 should read its own write: %v", readValueTxn1)
	}

	readValueTxn2, foundTxn2 := store.Read(key, txn2)
	if !foundTxn2 || readValueTxn2 != valueTxn2 {
		t.Errorf("Txn2 should read its own write: %v", readValueTxn2)
	}

	var txWg sync.WaitGroup
	txWg.Add(1)
	go func() {
		defer txWg.Done()
		//fmt.Println("Committing txn1")
		if !store.Commit(txn1) {
			t.Errorf("Failed to commit transaction1")
		}
	}()

	txWg.Add(1)
	go func() {
		defer txWg.Done()
		//fmt.Println("Committing txn2")
		if !store.Commit(txn2) {
			t.Errorf("Failed to commit transaction2")
		}
	}()

	txWg.Wait()

	_, _ = store.Read(key, int64(rand.Int()))
	//fmt.Println(readValueTxn3)
}
