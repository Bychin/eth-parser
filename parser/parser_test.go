package parser

import (
	"reflect"
	"testing"

	"eth-parser/eth"
)

type dummyTransactionsStorage map[string][]eth.Transaction

func (d *dummyTransactionsStorage) Init() error {
	return nil
}

func (d *dummyTransactionsStorage) Shutdown() error {
	return nil
}

func (d *dummyTransactionsStorage) Store(address string, transaction eth.Transaction) error {
	if _, ok := (*d)[address]; !ok {
		(*d)[address] = make([]eth.Transaction, 0)
	}

	(*d)[address] = append((*d)[address], transaction)
	return nil
}

func (d *dummyTransactionsStorage) Get(address string) ([]eth.Transaction, error) {
	return (*d)[address], nil
}

type dummyAddressesMapStorage map[string]struct{}

func (d *dummyAddressesMapStorage) Init() error {
	return nil
}

func (d *dummyAddressesMapStorage) Shutdown() error {
	return nil
}

func (d *dummyAddressesMapStorage) Store(address string) error {
	(*d)[address] = struct{}{}
	return nil
}

func (d *dummyAddressesMapStorage) Check(address string) bool {
	_, ok := (*d)[address]
	return ok
}

type dummyEthStream struct {
	blocks []*eth.Block
}

func (d *dummyEthStream) Init() error {
	return nil
}

func (d *dummyEthStream) Shutdown() error {
	return nil
}

func (d *dummyEthStream) Routine() {}

func (d *dummyEthStream) LastBlockNumber() int64 {
	return 0
}

func (d *dummyEthStream) BlocksQueue() <-chan *eth.Block {
	ch := make(chan *eth.Block, len(d.blocks))
	go func() {
		for _, b := range d.blocks {
			ch <- b
		}

		close(ch)
	}()
	return ch
}

func TestParse(t *testing.T) {
	blocks := []*eth.Block{
		{
			Number: "1",
			Transactions: []eth.Transaction{
				{
					Hash: "hash1",
					From: "from1",
					To:   "to1",
				},
				{
					Hash: "hash2",
					From: "from1",
					To:   "to1",
				},
				{
					Hash: "hash3",
					From: "from2",
					To:   "to1",
				},
			},
		},
		{
			Number:       "2",
			Transactions: nil,
		},
		{
			Number: "3",
			Transactions: []eth.Transaction{
				{
					Hash: "hash4",
					From: "from2",
					To:   "to3",
				},
				{
					Hash: "hash5",
					From: "from3",
					To:   "to3",
				},
				{
					Hash: "hash6",
					From: "from4",
					To:   "to4",
				},
			},
		},
	}

	ethPoller := &dummyEthStream{
		blocks: blocks,
	}

	addressesStorage := &dummyAddressesMapStorage{
		"from1": struct{}{},
		"from2": struct{}{},
		"to3":   struct{}{},
	}

	transactionsStorage := &dummyTransactionsStorage{}

	p := NewParser(ethPoller, transactionsStorage, addressesStorage)
	p.Routine()
	p.Shutdown()

	expectedTransactionsStorage := &dummyTransactionsStorage{
		"from1": []eth.Transaction{
			{
				Hash: "hash1",
				From: "from1",
				To:   "to1",
			},
			{
				Hash: "hash2",
				From: "from1",
				To:   "to1",
			},
		},
		"from2": []eth.Transaction{
			{
				Hash: "hash3",
				From: "from2",
				To:   "to1",
			},
			{
				Hash: "hash4",
				From: "from2",
				To:   "to3",
			},
		},
		"to3": []eth.Transaction{
			{
				Hash: "hash4",
				From: "from2",
				To:   "to3",
			},
			{
				Hash: "hash5",
				From: "from3",
				To:   "to3",
			},
		},
	}

	ok := reflect.DeepEqual(transactionsStorage, expectedTransactionsStorage)
	if !ok {
		t.Errorf("transaction storages are not equal:\nhave: %+v\nwant: %+v]",
			transactionsStorage,
			expectedTransactionsStorage,
		)
	}
}
