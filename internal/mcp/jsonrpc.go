package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
)

// JSON-RPC 2.0 error codes
const (
	// Standard JSON-RPC error codes
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603

	// MCP-specific error codes
	ServerError = -32000
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface
func (e *JSONRPCError) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("JSON-RPC error %d: %s (data: %v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// NewRequest creates a new JSON-RPC 2.0 request
func NewRequest(id interface{}, method string, params interface{}) *JSONRPCRequest {
	return &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

// NewNotification creates a new JSON-RPC 2.0 notification (no ID)
func NewNotification(method string, params interface{}) *JSONRPCRequest {
	return &JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

// NewResponse creates a new JSON-RPC 2.0 success response
func NewResponse(id interface{}, result interface{}) (*JSONRPCResponse, error) {
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  resultBytes,
	}, nil
}

// NewErrorResponse creates a new JSON-RPC 2.0 error response
func NewErrorResponse(id interface{}, code int, message string, data interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// IsError returns true if the response contains an error
func (r *JSONRPCResponse) IsError() bool {
	return r.Error != nil
}

// UnmarshalResult unmarshals the result into the provided value
func (r *JSONRPCResponse) UnmarshalResult(v interface{}) error {
	if r.IsError() {
		return r.Error
	}
	if r.Result == nil {
		return errors.New("result is nil")
	}
	return json.Unmarshal(r.Result, v)
}

// Validate validates a JSON-RPC request
func (r *JSONRPCRequest) Validate() error {
	if r.JSONRPC != "2.0" {
		return fmt.Errorf("invalid JSON-RPC version: %s", r.JSONRPC)
	}
	if r.Method == "" {
		return errors.New("method is required")
	}
	return nil
}

// Validate validates a JSON-RPC response
func (r *JSONRPCResponse) Validate() error {
	if r.JSONRPC != "2.0" {
		return fmt.Errorf("invalid JSON-RPC version: %s", r.JSONRPC)
	}
	if r.Result == nil && r.Error == nil {
		return errors.New("either result or error must be present")
	}
	if r.Result != nil && r.Error != nil {
		return errors.New("result and error cannot both be present")
	}
	return nil
}
