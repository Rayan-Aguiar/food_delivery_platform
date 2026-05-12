package broker

import "testing"

func TestNew_InvalidURL(t *testing.T) {
	_, err := New(Config{URL: "amqp://%zz"})
	if err == nil {
		t.Fatal("expected error for invalid amqp url")
	}
}

func TestRabbit_ChannelAndCloseNil(t *testing.T) {
	r := &Rabbit{}
	if r.Channel() != nil {
		t.Fatal("expected nil channel")
	}
	if err := r.Close(); err != nil {
		t.Fatalf("expected nil close error, got: %v", err)
	}
}
