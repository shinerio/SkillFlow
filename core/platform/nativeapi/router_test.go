package nativeapi

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouterReturnsMethodNotFoundForUnknownMethod(t *testing.T) {
	router := NewRouter()

	resp := router.Handle(context.Background(), Request{
		Version:   "2026-04-25",
		Method:    "skills.missing",
		Params:    json.RawMessage(`{}`),
		RequestID: "req-1",
	})

	assert.False(t, resp.OK)
	require.NotNil(t, resp.Error)
	assert.Equal(t, CodeMethodNotFound, resp.Error.Code)
	assert.Equal(t, MessageKeyMethodNotFound, resp.Error.MessageKey)
}

func TestRouterConvertsHandlerPanicToInternalError(t *testing.T) {
	router := NewRouter()
	router.Register("panic", func(context.Context, json.RawMessage) (any, error) {
		panic("handler failed")
	})

	resp := router.Handle(context.Background(), Request{
		Version:   "2026-04-25",
		Method:    "panic",
		Params:    json.RawMessage(`null`),
		RequestID: "req-2",
	})

	assert.False(t, resp.OK)
	require.NotNil(t, resp.Error)
	assert.Equal(t, CodeInternalError, resp.Error.Code)
	assert.Equal(t, MessageKeyInternalError, resp.Error.MessageKey)
	assert.Contains(t, resp.Error.Message, "handler failed")
}
