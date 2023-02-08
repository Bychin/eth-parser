package poller

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"eth-parser/eth"
	"eth-parser/jsonrpc"
)

const (
	methodEthBlockNumber      = "eth_blockNumber"
	methodEthGetBlockByNumber = "eth_getBlockByNumber"
)

var (
	ErrResourceNotFound = fmt.Errorf("resource not found")
)

type EthPoller struct {
	config *EthPollerConfig

	httpClient *http.Client
	reqID      uint

	initialBlockNumber int64
	lastBlockNumber    int64
	mu                 sync.RWMutex

	blocksQueue chan *eth.Block
	shutdown    chan struct{}
}

func NewEthPoller(config *EthPollerConfig) *EthPoller {
	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        config.MaxIdleConns,
			MaxConnsPerHost:     config.MaxConnsPerHost,
			MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		},
	}

	return &EthPoller{
		config:      config,
		httpClient:  httpClient,
		reqID:       0,
		mu:          sync.RWMutex{},
		blocksQueue: make(chan *eth.Block, config.QueueLen),
		shutdown:    make(chan struct{}),
	}
}

func (e *EthPoller) getBlockNumber() error {
	e.reqID++
	reqPacket := &jsonrpc.Packet{
		JSONRPC: jsonrpc.Version,
		ID:      e.reqID,
		Method:  methodEthBlockNumber,
	}

	respData, err := e.executePOSTRequestWithRetries(reqPacket)
	if err != nil {
		return fmt.Errorf("could not execute POST request: %w", err)
	}

	respPacket := &jsonrpc.Packet{}
	if unmarshalErr := json.Unmarshal(respData, respPacket); unmarshalErr != nil {
		return fmt.Errorf("could not unmarshal response packet: %w", unmarshalErr)
	}
	if respPacket.Error != nil {
		return fmt.Errorf("got response with error: %d, %s",
			respPacket.Error.Code, respPacket.Error.Message)
	}

	rawBlockNumber, ok := respPacket.Result.(string)
	if !ok {
		return fmt.Errorf("got wrong result type: %+v", respPacket.Result)
	}

	rawBlockNumber = strings.Replace(rawBlockNumber, "0x", "", -1)
	blockNumber, err := strconv.ParseInt(rawBlockNumber, 16, 64)
	if err != nil {
		return fmt.Errorf("could not parse block number from response: %w", err)
	}

	e.updateInitialBlockNumber(blockNumber)
	return nil
}

func (e *EthPoller) getBlockByNumber(number int64) error {
	numberAsStr := "0x" + strconv.FormatInt(number, 16)
	reqParams := []interface{}{numberAsStr, true}

	e.reqID++
	reqPacket := &jsonrpc.Packet{
		JSONRPC: jsonrpc.Version,
		ID:      e.reqID,
		Method:  methodEthGetBlockByNumber,
		Params:  reqParams,
	}

	respData, err := e.executePOSTRequestWithRetries(reqPacket)
	if err != nil {
		return fmt.Errorf("could not execute POST request: %w", err)
	}

	respPacket := &jsonrpc.Packet{
		Result: &eth.Block{},
	}
	if err := json.Unmarshal(respData, respPacket); err != nil {
		return fmt.Errorf("could not unmarshal response packet: %w", err)
	}
	if respPacket.Error != nil {
		code := respPacket.Error.Code
		if code == jsonrpc.CodeResourceNotFoundError {
			return ErrResourceNotFound
		}
		return fmt.Errorf("got response with error: %d, %s", code, respPacket.Error.Message)
	}

	block, ok := respPacket.Result.(*eth.Block)
	if !ok {
		return fmt.Errorf("got wrong result type: %+v", respPacket.Result)
	}

	e.blocksQueue <- block
	return nil
}

func (e *EthPoller) Init() error {
	log.Println("eth_poller: initializing")

	if err := e.getBlockNumber(); err != nil {
		return err
	}

	log.Printf("eth_poller: initial block #%d\n", e.initialBlockNumber)
	log.Println("eth_poller: successfully initialized")
	return nil
}

func (e *EthPoller) InitialBlockNumber() int64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.initialBlockNumber
}

func (e *EthPoller) updateInitialBlockNumber(newNumber int64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.initialBlockNumber = newNumber
}

func (e *EthPoller) LastBlockNumber() int64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.lastBlockNumber
}

func (e *EthPoller) updateLastBlockNumber(newNumber int64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.lastBlockNumber = newNumber
}

func (e *EthPoller) BlocksQueue() <-chan *eth.Block {
	return e.blocksQueue
}

func (e *EthPoller) Routine() {
	defer close(e.blocksQueue)

	if err := e.getBlockByNumber(e.initialBlockNumber); err != nil {
		log.Fatalf("eth_poller: could not start routine: %s", err)
	}

	e.updateLastBlockNumber(e.initialBlockNumber)

	for {
		select {
		case <-e.shutdown:
			return
		case <-time.After(e.config.PollInterval):
			// polling, pass
		}

		nextBlockNumber := e.lastBlockNumber + 1
		if err := e.getBlockByNumber(nextBlockNumber); err != nil {
			if errors.Is(err, ErrResourceNotFound) {
				log.Printf("eth_poller: waiting for block #%d", nextBlockNumber)
				continue
			}
			log.Printf("eth_poller: could not get block #%d: %s", nextBlockNumber, err)
			continue
		}

		e.updateLastBlockNumber(nextBlockNumber)
	}
}

func (e *EthPoller) Shutdown() error {
	log.Println("eth_poller: starting shutdown")

	close(e.shutdown)

	log.Println("eth_poller: successfully shutdown")
	return nil
}
