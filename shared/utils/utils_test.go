package utils

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	payload := map[string]string{"ok": "1"}

	if err := WriteJSON(rr, http.StatusAccepted, payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rr.Code != http.StatusAccepted {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected content type: %s", rr.Header().Get("Content-Type"))
	}

	var out map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if out["ok"] != "1" {
		t.Fatalf("unexpected payload: %+v", out)
	}
}

func TestPtrAndValOr(t *testing.T) {
	v := Ptr(10)
	if v == nil || *v != 10 {
		t.Fatalf("unexpected pointer value: %+v", v)
	}
	if got := ValOr(v, 20); got != 10 {
		t.Fatalf("unexpected ValOr non-nil: %d", got)
	}
	if got := ValOr[int](nil, 20); got != 20 {
		t.Fatalf("unexpected ValOr nil: %d", got)
	}
}

func TestStringFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "k", "v")
	if got := StringFromContext(ctx, "k"); got != "v" {
		t.Fatalf("unexpected value: %s", got)
	}
	if got := StringFromContext(ctx, "missing"); got != "" {
		t.Fatalf("expected empty for missing key, got: %s", got)
	}
	ctx2 := context.WithValue(context.Background(), "k2", 123)
	if got := StringFromContext(ctx2, "k2"); got != "" {
		t.Fatalf("expected empty for non-string value, got: %s", got)
	}
}
