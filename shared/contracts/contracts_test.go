package contracts

import (
	"encoding/json"
	"testing"
)

func TestEventJSONTags(t *testing.T) {
	e := Event[map[string]string]{
		Meta: EventMeta{
			EventID:       "1",
			EventType:     "x",
			EventVersion:  1,
			OccurredAt:    "now",
			Producer:      "svc",
			CorrelationID: "corr",
		},
		Payload: map[string]string{"k": "v"},
	}

	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := out["meta"]; !ok {
		t.Fatal("expected meta field")
	}
	if _, ok := out["payload"]; !ok {
		t.Fatal("expected payload field")
	}
}

func TestErrorResponseJSONTags(t *testing.T) {
	e := ErrorResponse{Code: "X", Message: "msg", RequestID: "r", CorrelationID: "c"}
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, k := range []string{"code", "message", "request_id", "correlation_id"} {
		if _, ok := out[k]; !ok {
			t.Fatalf("expected field %s", k)
		}
	}
}
