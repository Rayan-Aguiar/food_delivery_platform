package apperrors

import "net/http"

// AppError padroniza erros de aplicacao entre servicos.
type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Cause      error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func New(code, message string, statusCode int, cause error) *AppError {
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Cause:      cause,
	}
}
