package parser

import (
	"eth-parser/eth"
)

type transactionsStorage interface {
	Init() error
	Shutdown() error

	// Store stores transaction for address
	Store(address string, transaction eth.Transaction) error

	// Get returns stored transactions for address
	Get(address string) ([]eth.Transaction, error)
}

type addressesStorage interface {
	Init() error
	Shutdown() error

	// Store stores address
	Store(address string) error

	// Check checks if the address is saved
	Check(address string) bool
}
