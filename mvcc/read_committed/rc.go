package read_committed

import (
	"math/rand"
	"sync"
	"time"
)

type DataVersion struct {
	Value     interface{}
	Timestamp int64
}

type VersionedData struct {
	Versions []DataVersion
}

type MVCCStore struct {
	Data         map[string]*VersionedData
	Transactions map[int64]*Transaction
	lock         sync.Mutex
}

type Transaction struct {
	ID        int64
	StartTime int64
	Changes   map[string]DataVersion
	Status    string
}

func NewMVCCStore() *MVCCStore {
	return &MVCCStore{
		Data:         make(map[string]*VersionedData),
		Transactions: make(map[int64]*Transaction),
	}
}

func (store *MVCCStore) Begin() int64 {
	store.lock.Lock()
	defer store.lock.Unlock()

	transactionID := rand.Int63()
	store.Transactions[transactionID] = &Transaction{
		ID:      transactionID,
		Changes: make(map[string]DataVersion),
		Status:  "active",
	}
	return transactionID
}

func (store *MVCCStore) Rollback(transactionID int64) bool {
	store.lock.Lock()
	defer store.lock.Unlock()

	_, exists := store.Transactions[transactionID]
	if !exists || store.Transactions[transactionID].Status != "active" {
		return false
	}

	delete(store.Transactions, transactionID)
	return true
}

func (store *MVCCStore) Read(key string) (interface{}, bool) {
	store.lock.Lock()
	defer store.lock.Unlock()

	if versions, exists := store.Data[key]; exists {
		return versions.Versions[len(versions.Versions)-1].Value, true
	}

	return nil, false
}

func (store *MVCCStore) Write(key string, value interface{}, transactionID int64) bool {
	store.lock.Lock()
	defer store.lock.Unlock()

	transaction, exists := store.Transactions[transactionID]
	if !exists || transaction.Status != "active" {
		return false
	}

	change := DataVersion{
		Value: value,
	}
	transaction.Changes[key] = change
	return true
}

func (store *MVCCStore) Commit(transactionID int64) bool {
	store.lock.Lock()
	defer store.lock.Unlock()

	transaction, exists := store.Transactions[transactionID]
	if !exists || transaction.Status != "active" {
		return false
	}

	commitTime := time.Now().UnixNano()
	for key, change := range transaction.Changes {
		change.Timestamp = commitTime
		if _, exists := store.Data[key]; !exists {
			store.Data[key] = &VersionedData{}
		}
		store.Data[key].Versions = append(store.Data[key].Versions, change)
	}

	transaction.Status = "committed"
	delete(store.Transactions, transactionID)
	return true
}
