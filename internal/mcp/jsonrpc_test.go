package mcp

import (
	"encoding/json"
	"sync"
	"testing"
)

func TestNewRequest(t *testing.T) {
	tests := []struct {
		name   string
		id     interface{}
		method string
		params interface{}
		want   *JSONRPCRequest
	}{
		{
			name:   "request with string ID",
			id:     "test-123",
			method: "test_method",
			params: map[string]string{"key": "value"},
			want: &JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      "test-123",
				Method:  "test_method",
				Params:  map[string]string{"key": "value"},
			},
		},
		{
			name:   "request with numeric ID",
			id:     42,
			method: "initialize",
			params: nil,
			want: &JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      42,
				Method:  "initialize",
				Params:  nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRequest(tt.id, tt.method, tt.params)
			if got.JSONRPC != tt.want.JSONRPC {
				t.Errorf("JSONRPC = %v, want %v", got.JSONRPC, tt.want.JSONRPC)
			}
			if got.Method != tt.want.Method {
				t.Errorf("Method = %v, want %v", got.Method, tt.want.Method)
			}
			if got.ID != tt.want.ID {
				t.Errorf("ID = %v, want %v", got.ID, tt.want.ID)
			}
		})
	}
}

func TestNewNotification(t *testing.T) {
	method := "notification_method"
	params := map[string]interface{}{"event": "test"}

	notification := NewNotification(method, params)

	if notification.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %v, want 2.0", notification.JSONRPC)
	}
	if notification.Method != method {
		t.Errorf("Method = %v, want %v", notification.Method, method)
	}
	if notification.ID != nil {
		t.Errorf("ID should be nil for notifications, got %v", notification.ID)
	}
}

func TestRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request *JSONRPCRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "test_method",
				ID:      1,
			},
			wantErr: false,
		},
		{
			name: "invalid version",
			request: &JSONRPCRequest{
				JSONRPC: "1.0",
				Method:  "test_method",
				ID:      1,
			},
			wantErr: true,
		},
		{
			name: "missing method",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "",
				ID:      1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequestResponseMarshaling(t *testing.T) {
	// Test request marshaling
	req := NewRequest(123, "tools/list", map[string]interface{}{
		"limit":  10,
		"offset": 0,
	})

	reqBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Test request unmarshaling
	var unmarshaledReq JSONRPCRequest
	if err := json.Unmarshal(reqBytes, &unmarshaledReq); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if unmarshaledReq.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %v, want 2.0", unmarshaledReq.JSONRPC)
	}
	if unmarshaledReq.Method != "tools/list" {
		t.Errorf("Method = %v, want tools/list", unmarshaledReq.Method)
	}

	// Test response marshaling
	result := map[string]interface{}{
		"tools": []string{"tool1", "tool2"},
	}
	resp, err := NewResponse(123, result)
	if err != nil {
		t.Fatalf("Failed to create response: %v", err)
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Test response unmarshaling
	var unmarshaledResp JSONRPCResponse
	if err := json.Unmarshal(respBytes, &unmarshaledResp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if unmarshaledResp.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %v, want 2.0", unmarshaledResp.JSONRPC)
	}

	// Verify result can be unmarshaled
	var resultData map[string]interface{}
	if err := unmarshaledResp.UnmarshalResult(&resultData); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	tools, ok := resultData["tools"].([]interface{})
	if !ok || len(tools) != 2 {
		t.Errorf("Expected 2 tools in result, got %v", resultData)
	}
}

func TestErrorResponse(t *testing.T) {
	errResp := NewErrorResponse(456, InvalidParams, "Invalid parameters provided", map[string]string{
		"param": "limit",
		"issue": "must be positive",
	})

	if !errResp.IsError() {
		t.Error("IsError() should return true for error response")
	}

	if errResp.Error.Code != InvalidParams {
		t.Errorf("Error code = %d, want %d", errResp.Error.Code, InvalidParams)
	}

	if errResp.Error.Message != "Invalid parameters provided" {
		t.Errorf("Error message = %s, want 'Invalid parameters provided'", errResp.Error.Message)
	}

	// Test error string formatting
	errStr := errResp.Error.Error()
	if errStr == "" {
		t.Error("Error() should return non-empty string")
	}
}

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		name string
		code int
		want int
	}{
		{"ParseError", ParseError, -32700},
		{"InvalidRequest", InvalidRequest, -32600},
		{"MethodNotFound", MethodNotFound, -32601},
		{"InvalidParams", InvalidParams, -32602},
		{"InternalError", InternalError, -32603},
		{"ServerError", ServerError, -32000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.code, tt.want)
			}
		})
	}
}

func TestResponseValidate(t *testing.T) {
	tests := []struct {
		name     string
		response *JSONRPCResponse
		wantErr  bool
	}{
		{
			name: "valid success response",
			response: &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
				Result:  json.RawMessage(`{"status":"ok"}`),
			},
			wantErr: false,
		},
		{
			name: "valid error response",
			response: &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
				Error: &JSONRPCError{
					Code:    InternalError,
					Message: "Internal error",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid version",
			response: &JSONRPCResponse{
				JSONRPC: "1.0",
				ID:      1,
				Result:  json.RawMessage(`{}`),
			},
			wantErr: true,
		},
		{
			name: "missing result and error",
			response: &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
			},
			wantErr: true,
		},
		{
			name: "both result and error present",
			response: &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
				Result:  json.RawMessage(`{}`),
				Error:   &JSONRPCError{Code: InternalError, Message: "error"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.response.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnmarshalResult(t *testing.T) {
	tests := []struct {
		name    string
		resp    *JSONRPCResponse
		want    interface{}
		wantErr bool
	}{
		{
			name: "successful unmarshal",
			resp: &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
				Result:  json.RawMessage(`{"count":42}`),
			},
			want:    &struct{ Count int }{Count: 42},
			wantErr: false,
		},
		{
			name: "error response",
			resp: &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
				Error:   &JSONRPCError{Code: InternalError, Message: "error"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "nil result",
			resp: &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
				Result:  nil,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result struct{ Count int }
			err := tt.resp.UnmarshalResult(&result)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalResult() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.want != nil {
				want := tt.want.(*struct{ Count int })
				if result.Count != want.Count {
					t.Errorf("UnmarshalResult() result = %v, want %v", result, want)
				}
			}
		})
	}
}

func TestIDMatching(t *testing.T) {
	// Test that request and response IDs can be matched
	requestID := "req-12345"
	req := NewRequest(requestID, "test_method", nil)

	resp, err := NewResponse(requestID, map[string]string{"status": "ok"})
	if err != nil {
		t.Fatalf("Failed to create response: %v", err)
	}

	if req.ID != resp.ID {
		t.Errorf("Request ID %v does not match response ID %v", req.ID, resp.ID)
	}

	// Test with numeric ID
	numericID := 999
	req2 := NewRequest(numericID, "another_method", nil)
	resp2, _ := NewResponse(numericID, nil)

	// Type assertion needed for interface{} comparison
	reqIDFloat, ok1 := req2.ID.(int)
	respIDFloat, ok2 := resp2.ID.(int)

	if !ok1 || !ok2 {
		t.Fatal("Failed to convert IDs to comparable types")
	}

	if reqIDFloat != respIDFloat {
		t.Errorf("Request ID %v does not match response ID %v", req2.ID, resp2.ID)
	}
}

func TestConcurrentRequestHandling(t *testing.T) {
	// Test that multiple requests can be created and matched concurrently
	const numRequests = 100
	var wg sync.WaitGroup
	wg.Add(numRequests)

	results := make(map[int]bool)
	var mu sync.Mutex

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			defer wg.Done()

			// Create request
			req := NewRequest(id, "concurrent_method", map[string]int{"value": id})

			// Create response
			resp, err := NewResponse(id, map[string]int{"result": id * 2})
			if err != nil {
				t.Errorf("Failed to create response for ID %d: %v", id, err)
				return
			}

			// Verify ID matching
			if req.ID != resp.ID {
				t.Errorf("ID mismatch: request %v != response %v", req.ID, resp.ID)
				return
			}

			// Record successful match
			mu.Lock()
			results[id] = true
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify all requests were processed
	if len(results) != numRequests {
		t.Errorf("Expected %d successful matches, got %d", numRequests, len(results))
	}
}
