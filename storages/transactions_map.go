package storages

import (
	"fmt"
	"sync"

	"eth-parser/eth"
)

const (
	initialStorageCap                = 1024
	initialTransactionsPerAddressCap = 1024
)

var (
	ErrUninitialized = fmt.Errorf("storage is uninitialized")
	ErrInternal      = fmt.Errorf("internal error")
)

type TransactionsMapStorage struct {
	storage       map[string][]eth.Transaction
	storageMu     sync.RWMutex
	resetAfterGet bool
}

func NewTransactionsMapStorage(resetAfterGet bool) *TransactionsMapStorage {
	return &TransactionsMapStorage{
		storage:       make(map[string][]eth.Transaction, initialStorageCap),
		storageMu:     sync.RWMutex{},
		resetAfterGet: resetAfterGet,
	}
}

func (m *TransactionsMapStorage) Init() error {
	return nil
}

func (m *TransactionsMapStorage) Shutdown() error {
	return nil
}

func (m *TransactionsMapStorage) Store(address string, transaction eth.Transaction) error {
	if m == nil {
		return ErrUninitialized
	}

	m.storageMu.Lock()
	defer m.storageMu.Unlock()

	if _, ok := m.storage[address]; !ok {
		m.storage[address] = make([]eth.Transaction, 0, initialTransactionsPerAddressCap)
	}

	m.storage[address] = append(m.storage[address], transaction)
	return nil
}

func (m *TransactionsMapStorage) Get(address string) ([]eth.Transaction, error) {
	if m == nil {
		return nil, ErrUninitialized
	}

	m.storageMu.RLock()
	result := m.storage[address]
	m.storageMu.RUnlock()

	if m.resetAfterGet {
		m.storageMu.Lock()
		delete(m.storage, address)
		m.storageMu.Unlock()
	}

	return result, nil
}
