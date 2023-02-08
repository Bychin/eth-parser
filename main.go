package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"eth-parser/parser"
	"eth-parser/poller"
	"eth-parser/server"
	"eth-parser/storages"
)

const (
	cloudflareEndpoint = "https://cloudflare-eth.com"

	serverShutdownTimeout = 5 * time.Second

	defaultServerAddr                = "localhost:8080"
	defaultStorageReset              = false
	defaultPollerEndpoint            = cloudflareEndpoint
	defaultPollerPollInterval        = 1 * time.Second
	defaultPollerTimeout             = 5 * time.Second
	defaultPollerMaxIdleConns        = 100
	defaultPollerMaxConnsPerHost     = 100
	defaultPollerMaxIdleConnsPerHost = 100
	defaultPollerNumRetries          = 3
	defaultPollerQueueLen            = 10
)

var (
	serverAddr = flag.String("server.addr",
		defaultServerAddr, "server addr to listen on")

	storageReset = flag.Bool("storage.reset",
		defaultStorageReset, "reset stored transactions after getting them")

	pollerEndpoint = flag.String("poller.endpoint",
		defaultPollerEndpoint, "endpoint for poller to get info about ETH blocks")
	pollerPollInterval = flag.Duration("poller.interval",
		defaultPollerPollInterval, "poll interval")
	pollerTimeout = flag.Duration("poller.timeout",
		defaultPollerTimeout, "endpoint request timeout")
	pollerMaxIdleConns = flag.Int("poller.max_idle_conns",
		defaultPollerMaxIdleConns, "max idle conns")
	pollerMaxConnsPerHost = flag.Int("poller.max_conns_per_host",
		defaultPollerMaxConnsPerHost, "max conns per host")
	pollerMaxIdleConnsPerHost = flag.Int("poller.max_idle_conns_per_host",
		defaultPollerMaxIdleConnsPerHost, "max idle conns per host")
	pollerNumRetries = flag.Int("poller.num_retries",
		defaultPollerNumRetries, "num retries")
	pollerQueueLen = flag.Int("poller.queue_len",
		defaultPollerQueueLen, "queue length")
)

func main() {
	flag.Parse()

	pollerConfig := &poller.EthPollerConfig{
		Endpoint:            *pollerEndpoint,
		PollInterval:        *pollerPollInterval,
		Timeout:             *pollerTimeout,
		MaxIdleConns:        *pollerMaxIdleConns,
		MaxConnsPerHost:     *pollerMaxConnsPerHost,
		MaxIdleConnsPerHost: *pollerMaxIdleConnsPerHost,
		NumRetries:          *pollerNumRetries,
		QueueLen:            *pollerQueueLen,
	}

	ethPoller := poller.NewEthPoller(pollerConfig)
	transactionsStorage := storages.NewTransactionsMapStorage(*storageReset)
	addressesStorage := storages.NewAddressesMapStorage()

	p := parser.NewParser(ethPoller, transactionsStorage, addressesStorage)

	if err := p.Init(); err != nil {
		log.Fatalf("main: could not init parser: %s", err)
	}

	go p.Routine()
	defer p.Shutdown()

	log.Printf("main: starting HTTP server")
	httpServer, httpServerExit := server.InitHTTPServer(server.NewHandler(p), *serverAddr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Printf("main: got SIGINT")

	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("main: could not shutdown server successfully: %s", err)
	}

	<-httpServerExit
}
