package parser

import (
	"fmt"
	"log"

	"eth-parser/eth"
)

type Parser struct {
	ethStream     ethStream
	transactions  transactionsStorage
	subscriptions addressesStorage

	shutdown chan struct{}
}

func NewParser(
	ethStream ethStream,
	transactions transactionsStorage,
	subscriptions addressesStorage,
) *Parser {
	return &Parser{
		ethStream:     ethStream,
		transactions:  transactions,
		subscriptions: subscriptions,
		shutdown:      make(chan struct{}),
	}
}

func (p *Parser) Init() error {
	log.Println("parser: initializing")

	if err := p.ethStream.Init(); err != nil {
		return fmt.Errorf("could not initialize ETH stream: %w", err)
	}
	if err := p.transactions.Init(); err != nil {
		return fmt.Errorf("could not initialize transactions storage: %w", err)
	}
	if err := p.subscriptions.Init(); err != nil {
		return fmt.Errorf("could not initialize subscriptions storage: %w", err)
	}

	log.Println("parser: successfully initialized")
	return nil
}

func (p *Parser) Shutdown() {
	log.Println("parser: starting shutdown")

	if err := p.ethStream.Shutdown(); err != nil {
		log.Printf("parser: got err on ETH stream shutdown: %s", err)
	}
	<-p.shutdown

	if err := p.transactions.Shutdown(); err != nil {
		log.Printf("parser: got err on transactions storage shutdown: %s", err)
	}
	if err := p.subscriptions.Shutdown(); err != nil {
		log.Printf("parser: got err on subscriptions storage shutdown: %s", err)
	}

	log.Println("parser: successfully shutdown")
}

func (p *Parser) Routine() {
	go p.ethStream.Routine()

	for block := range p.ethStream.BlocksQueue() {
		log.Printf("parser: got next block '%s' with %d transactions\n",
			block.Number, len(block.Transactions))

		for _, transaction := range block.Transactions {
			for _, addr := range []string{transaction.From, transaction.To} {
				if ok := p.subscriptions.Check(addr); !ok {
					continue
				}
				if err := p.transactions.Store(addr, transaction); err != nil {
					log.Printf("parser: could not store transaction for '%s' [%+v]: %s",
						addr, transaction, err)
				}

				log.Printf("parser: stored transaction for '%s' [%+v]", addr, transaction)
			}
		}
	}

	close(p.shutdown)
}

func (p *Parser) GetCurrentBlock() int {
	if p == nil || p.ethStream == nil {
		return -1
	}

	return int(p.ethStream.LastBlockNumber())
}

func (p *Parser) Subscribe(address string) bool {
	if p == nil || p.subscriptions == nil {
		return false
	}

	if err := p.subscriptions.Store(address); err != nil {
		log.Printf("parser: could not store subscription for '%s'", address)
		return false
	}

	log.Printf("parser: subscribed '%s' successfully", address)
	return true
}

func (p *Parser) GetTransactions(address string) []eth.Transaction {
	if p == nil || p.transactions == nil {
		return nil
	}

	result, err := p.transactions.Get(address)
	if err != nil {
		log.Printf("parser: could not get transactions: %s", err)
		return nil
	}

	return result
}
