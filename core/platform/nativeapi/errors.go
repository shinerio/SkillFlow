package nativeapi

const (
	CodeMethodNotFound = "method_not_found"
	CodeInternalError  = "internal_error"

	MessageKeyMethodNotFound = "nativeapi.error.method_not_found"
	MessageKeyInternalError  = "nativeapi.error.internal_error"
)

// Error carries a stable daemon API error code and optional display metadata.
type Error struct {
	Code       string `json:"code"`
	Message    string `json:"message,omitempty"`
	MessageKey string `json:"messageKey,omitempty"`
}

func errorResponse(code, message, messageKey string) Response {
	return Response{
		OK:     false,
		Result: nullResult(),
		Error: &Error{
			Code:       code,
			Message:    message,
			MessageKey: messageKey,
		},
	}
}
