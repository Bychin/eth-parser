package poller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"

	"eth-parser/jsonrpc"
)

var (
	errBadHTTPStatusCode = fmt.Errorf("bad http status code")
)

func (e *EthPoller) executePOSTRequestWithRetries(
	packet *jsonrpc.Packet,
) (
	respData []byte,
	err error,
) {
	data, marshalErr := json.Marshal(packet)
	if marshalErr != nil {
		return nil, fmt.Errorf("could not marshal packet: %w", marshalErr)
	}

	for i := 0; i < e.config.NumRetries; i++ {
		respData, err = e.executePOSTRequest(data)
		if err == nil {
			return
		}
		if !e.shouldRetry(err) {
			return nil, err
		}

		log.Printf(
			"eth_poller: executePOSTRequestWithRetries(): error on executing POST request [%d/%d]: %s",
			i+1, e.config.NumRetries, err,
		)
		if i < e.config.NumRetries-1 {
			log.Printf("eth_poller: executePOSTRequestWithRetries(): will retry")
		}
	}

	return nil, err
}

func (e *EthPoller) executePOSTRequest(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(data)

	resp, err := e.httpClient.Post(e.config.Endpoint, "application/json", buf)
	if err != nil {
		return nil, fmt.Errorf("could not make request: %w", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, fmt.Errorf("got %w: %d", errBadHTTPStatusCode, resp.StatusCode)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("got unexpected http status code: %d", resp.StatusCode)
	}

	return body, nil
}

func (e *EthPoller) shouldRetry(httpErr error) bool {
	if errors.Is(httpErr, errBadHTTPStatusCode) {
		return true
	}
	var netErr net.Error
	if errors.As(httpErr, &netErr) && netErr.Timeout() {
		return true
	}
	syscallErr := &os.SyscallError{}
	if errors.As(httpErr, &syscallErr) && errors.Is(syscallErr.Err, syscall.ECONNRESET) {
		return true
	}

	return false
}
