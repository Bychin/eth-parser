package server

import (
	"errors"
	"log"
	"net/http"
)

func InitHTTPServer(h *Handler, addr string) (*http.Server, <-chan struct{}) {
	if h == nil {
		log.Fatalf("server: got nil handler")
	}

	server := &http.Server{
		Addr: addr,
	}

	http.HandleFunc("/current_block", h.currentBlockHandler)
	http.HandleFunc("/subscribe", h.subscribeHandler)
	http.HandleFunc("/transactions", h.transactionsHandler)

	done := make(chan struct{})
	go func() {
		defer close(done)

		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: error on listen and serve: %s", err)
		}

		log.Println("server: shutdown successfully")
	}()

	return server, done
}
