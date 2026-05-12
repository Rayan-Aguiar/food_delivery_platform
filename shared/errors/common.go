package apperrors

import "net/http"

const (
	CodeInvalidArgument = "INVALID_ARGUMENT"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeForbidden       = "FORBIDDEN"
	CodeNotFound        = "NOT_FOUND"
	CodeConflict        = "CONFLICT"
	CodeInternal        = "INTERNAL"
)

func InvalidArgument(message string, cause error) *AppError {
	return New(CodeInvalidArgument, message, http.StatusBadRequest, cause)
}

func Unauthorized(message string, cause error) *AppError {
	return New(CodeUnauthorized, message, http.StatusUnauthorized, cause)
}

func Forbidden(message string, cause error) *AppError {
	return New(CodeForbidden, message, http.StatusForbidden, cause)
}

func NotFound(message string, cause error) *AppError {
	return New(CodeNotFound, message, http.StatusNotFound, cause)
}

func Conflict(message string, cause error) *AppError {
	return New(CodeConflict, message, http.StatusConflict, cause)
}

func Internal(message string, cause error) *AppError {
	return New(CodeInternal, message, http.StatusInternalServerError, cause)
}
