package nativeapi

import "encoding/json"

// Request is the stable JSON envelope accepted by native daemon API handlers.
type Request struct {
	Version   string          `json:"version"`
	Method    string          `json:"method"`
	Params    json.RawMessage `json:"params"`
	RequestID string          `json:"requestID"`
}

// Response is the stable JSON envelope returned by native daemon API handlers.
type Response struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result"`
	Error  *Error          `json:"error"`
}
