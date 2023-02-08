package jsonrpc

const (
	Version = "2.0"

	CodeResourceNotFoundError = -32001 // cloudflare custom error
)

type Packet struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      uint        `json:"id"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
