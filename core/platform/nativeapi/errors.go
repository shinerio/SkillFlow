package nativeapi

// Error carries a stable daemon API error code and optional display metadata.
type Error struct {
	Code       string `json:"code"`
	Message    string `json:"message,omitempty"`
	MessageKey string `json:"messageKey,omitempty"`
}
