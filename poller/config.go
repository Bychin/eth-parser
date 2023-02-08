package poller

import "time"

type EthPollerConfig struct {
	Endpoint     string
	PollInterval time.Duration

	Timeout             time.Duration
	MaxIdleConns        int
	MaxConnsPerHost     int
	MaxIdleConnsPerHost int
	NumRetries          int

	QueueLen int
}
