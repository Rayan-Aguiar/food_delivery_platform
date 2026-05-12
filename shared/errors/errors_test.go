package apperrors

import (
	"errors"
	"net/http"
	"testing"
)

func TestAppErrorBasics(t *testing.T) {
	cause := errors.New("db fail")
	err := New("X", "msg", 0, cause)

	if err.StatusCode != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", err.StatusCode)
	}
	if err.Error() != "msg" {
		t.Fatalf("unexpected error message: %s", err.Error())
	}
	if !errors.Is(err, cause) {
		t.Fatal("expected unwrap to expose cause")
	}

	var nilErr *AppError
	if nilErr.Error() != "" {
		t.Fatal("nil app error should return empty message")
	}
}

func TestCommonConstructors(t *testing.T) {
	tests := []struct {
		err    *AppError
		code   string
		status int
	}{
		{InvalidArgument("x", nil), CodeInvalidArgument, http.StatusBadRequest},
		{Unauthorized("x", nil), CodeUnauthorized, http.StatusUnauthorized},
		{Forbidden("x", nil), CodeForbidden, http.StatusForbidden},
		{NotFound("x", nil), CodeNotFound, http.StatusNotFound},
		{Conflict("x", nil), CodeConflict, http.StatusConflict},
		{Internal("x", nil), CodeInternal, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		if tt.err.Code != tt.code || tt.err.StatusCode != tt.status {
			t.Fatalf("unexpected constructor output: %+v", tt.err)
		}
	}
}

func TestToHTTPResponse(t *testing.T) {
	status, body := ToHTTPResponse(NotFound("missing", nil), "req-1", "corr-1")
	if status != http.StatusNotFound {
		t.Fatalf("unexpected status: %d", status)
	}
	if body.Code != CodeNotFound || body.Message != "missing" {
		t.Fatalf("unexpected body: %+v", body)
	}

	status, body = ToHTTPResponse(errors.New("unknown"), "req-2", "corr-2")
	if status != http.StatusInternalServerError {
		t.Fatalf("unexpected fallback status: %d", status)
	}
	if body.Code != CodeInternal || body.Message != "internal server error" {
		t.Fatalf("unexpected fallback body: %+v", body)
	}
}
