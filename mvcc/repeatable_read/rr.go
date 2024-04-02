package repeatable_read

import (
	"math/rand"
	"sync"
	"time"
)

type (
	// DataVersion represents a single version of a value.
	DataVersion struct {
		Value     interface{}
		Timestamp int64 // Timestamp of the version
	}

	// VersionedData holds all versions of a value sorted by timestamp.
	VersionedData struct {
		Versions []DataVersion
	}

	MVCCStore struct {
		Data         map[string]*VersionedData
		Transactions map[int64]*Transaction
		lock         sync.Mutex
	}

	Transaction struct {
		ID        int64
		StartTime int64
		Changes   map[string]interface{}
		Status    string
	}
)

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
		ID:        transactionID,
		StartTime: time.Now().UnixNano(),
		Changes:   make(map[string]interface{}),
		Status:    "active",
	}
	return transactionID
}

func (store *MVCCStore) Rollback(transactionID int64) bool {
	store.lock.Lock()
	defer store.lock.Unlock()

	transaction, exists := store.Transactions[transactionID]
	if !exists || transaction.Status != "active" {
		return false
	}

	transaction.Status = "aborted"
	delete(store.Transactions, transactionID)
	return true
}

func (store *MVCCStore) Commit(transactionID int64) bool {
	store.lock.Lock()
	defer store.lock.Unlock()

	transaction, exists := store.Transactions[transactionID]
	if !exists || transaction.Status != "active" {
		return false
	}

	for key, value := range transaction.Changes {
		if _, exists := store.Data[key]; !exists {
			store.Data[key] = &VersionedData{}
		}
		version := DataVersion{
			Timestamp: time.Now().UnixNano(),
			Value:     value,
		}
		store.Data[key].Versions = append(store.Data[key].Versions, version)
	}

	transaction.Status = "committed"
	delete(store.Transactions, transactionID)
	return true
}

func (store *MVCCStore) Read(key string, transactionID int64) (interface{}, bool) {
	store.lock.Lock()
	defer store.lock.Unlock()

	var readTime int64
	if transaction, exists := store.Transactions[transactionID]; exists {
		readTime = transaction.StartTime
		if data, exists := store.Transactions[transactionID].Changes[key]; exists {
			return data, true
		}
	} else {
		readTime = time.Now().UnixNano()
		if data, exists := store.Data[key]; exists {
			for i := len(data.Versions) - 1; i >= 0; i-- {
				version := data.Versions[i]
				if version.Timestamp <= readTime {
					return version.Value, true
				}
			}
		}
	}

	return nil, false
}

func (store *MVCCStore) Write(key string, value interface{}, transactionID int64) bool {
	store.lock.Lock()
	defer store.lock.Unlock()

	transaction, exists := store.Transactions[transactionID]
	if !exists {
		return false
	}

	transaction.Changes[key] = value
	return true
}

// GC removes versions older than the specified duration.
func (store *MVCCStore) GC(duration time.Duration) {
	store.lock.Lock()
	defer store.lock.Unlock()

	for key, versions := range store.Data {
		cutoff := time.Now().Add(-duration)
		var i int
		for i = len(versions.Versions) - 1; i >= 0; i-- {
			if time.UnixMicro(versions.Versions[i].Timestamp).Before(cutoff) {
				break
			}
		}
		store.Data[key].Versions = versions.Versions[i+1:]
	}
}
