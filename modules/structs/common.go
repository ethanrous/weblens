package structs

// WeblensErrorInfo represents an error response for API endpoints.
type WeblensErrorInfo struct {
	Error string `json:"error"`
} // @name WeblensErrorInfo

// WLResponseInfo represents a generic response message for API endpoints.
type WLResponseInfo struct {
	Message string `json:"message"`
} // @name WLResponseInfo
