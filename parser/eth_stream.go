package parser

import (
	"eth-parser/eth"
)

type ethStream interface {
	Init() error
	Shutdown() error

	// Routine starts getting new ETH blocks
	Routine()

	// BlocksQueue returns stream of parsed ETH blocks
	BlocksQueue() <-chan *eth.Block

	// LastBlockNumber returns number of last parsed block
	LastBlockNumber() int64
}
