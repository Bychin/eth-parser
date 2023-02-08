package storages

import (
	"sync"
)

type AddressesMapStorage struct {
	storage   map[string]struct{}
	storageMu sync.RWMutex
}

func NewAddressesMapStorage() *AddressesMapStorage {
	return &AddressesMapStorage{
		storage:   make(map[string]struct{}, initialStorageCap),
		storageMu: sync.RWMutex{},
	}
}

func (m *AddressesMapStorage) Init() error {
	return nil
}

func (m *AddressesMapStorage) Shutdown() error {
	return nil
}

func (m *AddressesMapStorage) Store(address string) error {
	if m == nil {
		return ErrUninitialized
	}

	m.storageMu.Lock()
	defer m.storageMu.Unlock()

	m.storage[address] = struct{}{}
	return nil
}

func (m *AddressesMapStorage) Check(address string) bool {
	if m == nil {
		return false
	}

	m.storageMu.RLock()
	defer m.storageMu.RUnlock()

	_, ok := m.storage[address]
	return ok
}
