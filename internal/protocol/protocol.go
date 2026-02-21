package protocol

import "encoding/json"

// CurrentVersion is the protocol version used by client and server.
const CurrentVersion = 1

// Request is the envelope sent from client to server over the socket.
// Each request specifies an action and a payload specific to that action.
type Request struct {
	Version int             `json:"version"`
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Response is the envelope sent from server to client.
type Response struct {
	Status string `json:"status"`
	Data   any    `json:"data,omitempty"`
	Error  string `json:"error,omitempty"`
}

// OkResponse creates a success response with optional data.
func OkResponse(data any) *Response {
	return &Response{Status: "ok", Data: data}
}

// ErrorResponse creates an error response with a human-readable message.
func ErrorResponse(msg string) *Response {
	return &Response{Status: "error", Error: msg}
}
