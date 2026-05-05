package nativeapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type Handler func(context.Context, json.RawMessage) (any, error)

type Router struct {
	handlers map[string]Handler
}

func NewRouter() *Router {
	return &Router{
		handlers: map[string]Handler{},
	}
}

func (r *Router) Register(method string, handler Handler) {
	if r == nil || handler == nil {
		return
	}
	if r.handlers == nil {
		r.handlers = map[string]Handler{}
	}
	r.handlers[strings.TrimSpace(method)] = handler
}

func (r *Router) Handle(ctx context.Context, req Request) (resp Response) {
	defer func() {
		if recovered := recover(); recovered != nil {
			resp = errorResponse(CodeInternalError, fmt.Sprint(recovered), MessageKeyInternalError)
		}
	}()

	method := strings.TrimSpace(req.Method)
	if r == nil || r.handlers == nil {
		return errorResponse(CodeMethodNotFound, "method not found", MessageKeyMethodNotFound)
	}
	handler, ok := r.handlers[method]
	if !ok {
		return errorResponse(CodeMethodNotFound, "method not found", MessageKeyMethodNotFound)
	}

	result, err := handler(ctx, req.Params)
	if err != nil {
		return errorResponse(CodeInternalError, err.Error(), MessageKeyInternalError)
	}
	payload, err := marshalResult(result)
	if err != nil {
		return errorResponse(CodeInternalError, err.Error(), MessageKeyInternalError)
	}
	return Response{
		OK:     true,
		Result: payload,
		Error:  nil,
	}
}

func marshalResult(result any) (json.RawMessage, error) {
	if result == nil {
		return nullResult(), nil
	}
	if payload, ok := result.(json.RawMessage); ok {
		if len(payload) == 0 {
			return nullResult(), nil
		}
		return payload, nil
	}
	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func nullResult() json.RawMessage {
	return json.RawMessage(`null`)
}
