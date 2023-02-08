package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"eth-parser/parser"
)

type Handler struct {
	parser *parser.Parser
}

func NewHandler(parser *parser.Parser) *Handler {
	return &Handler{
		parser: parser,
	}
}

func (h *Handler) currentBlockHandler(w http.ResponseWriter, r *http.Request) {
	_, err := io.WriteString(w, strconv.FormatInt(int64(h.parser.GetCurrentBlock()), 10))
	if err != nil {
		log.Printf("http_handler: could not write current block\n")
	}
}

func (h *Handler) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	address := r.FormValue("address")
	if len(address) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if ok := h.parser.Subscribe(address); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("http_handler: subscribed for '%s' successfully\n", address)
}

func (h *Handler) transactionsHandler(w http.ResponseWriter, r *http.Request) {
	address := r.FormValue("address")
	if len(address) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	transactions := h.parser.GetTransactions(address)
	if len(transactions) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	log.Printf("http_handler: got %d transactions for '%s'\n", len(transactions), address)

	data, err := json.Marshal(transactions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("http_handler: could not write transactions for '%s'\n", address)
	}
}
