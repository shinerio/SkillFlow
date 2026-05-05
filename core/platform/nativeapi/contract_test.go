package nativeapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContractEnvelope(t *testing.T) {
	rawRequest := []byte(`{
		"version": "2026-04-25",
		"method": "skills.list",
		"params": {"category": "tools"},
		"requestID": "req-1"
	}`)

	var req Request
	require.NoError(t, json.Unmarshal(rawRequest, &req))
	assert.Equal(t, "2026-04-25", req.Version)
	assert.Equal(t, "skills.list", req.Method)
	assert.Equal(t, "req-1", req.RequestID)
	assert.JSONEq(t, `{"category":"tools"}`, string(req.Params))

	requestPayload, err := json.Marshal(req)
	require.NoError(t, err)
	assertEnvelopeKeys(t, requestPayload, "version", "method", "params", "requestID")

	resp := Response{
		OK:     false,
		Result: json.RawMessage(`null`),
		Error: &Error{
			Code:       "method_not_found",
			MessageKey: "nativeapi.error.method_not_found",
		},
	}
	responsePayload, err := json.Marshal(resp)
	require.NoError(t, err)
	assertEnvelopeKeys(t, responsePayload, "ok", "result", "error")

	var decoded Response
	require.NoError(t, json.Unmarshal(responsePayload, &decoded))
	require.NotNil(t, decoded.Error)
	assert.False(t, decoded.OK)
	assert.Equal(t, "method_not_found", decoded.Error.Code)
	assert.Equal(t, "nativeapi.error.method_not_found", decoded.Error.MessageKey)
}

func assertEnvelopeKeys(t *testing.T, payload []byte, keys ...string) {
	t.Helper()

	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(payload, &fields))
	for _, key := range keys {
		assert.Contains(t, fields, key)
	}
}
