package contracts

type ErrorResponse struct {
    Code          string `json:"code"`
    Message       string `json:"message"`
    RequestID     string `json:"request_id,omitempty"`
    CorrelationID string `json:"correlation_id,omitempty"`
}