package httpdelivery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"food_delivery_platform/shared/contracts"

	"golang.org/x/time/rate"
)

func TestIPRateLimiter_AllowsWithinBurst(t *testing.T) {
	// burst=3: primeiras 3 requisições devem passar
	rl := newIPRateLimiter(rate.Every(time.Hour), 3)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := range 3 {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		r.RemoteAddr = "10.0.0.1:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("requisicao %d: status = %d, want 200", i+1, w.Code)
		}
	}
}

func TestIPRateLimiter_RejectsAfterBurst(t *testing.T) {
	// burst=2, rate muito lenta: 3ª requisição deve retornar 429
	rl := newIPRateLimiter(rate.Every(time.Hour), 2)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ip := "10.0.0.2:5678"
	for i := range 2 {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		r.RemoteAddr = ip
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("requisicao %d deveria passar, status = %d", i+1, w.Code)
		}
	}

	// 3ª requisição deve ser bloqueada
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.RemoteAddr = ip
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", w.Code)
	}
	var resp contracts.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode ErrorResponse: %v", err)
	}
	if resp.Code != codeTooManyRequests {
		t.Errorf("code = %q, want %q", resp.Code, codeTooManyRequests)
	}
}

func TestIPRateLimiter_IndependentPerIP(t *testing.T) {
	// IPs diferentes têm limites independentes
	rl := newIPRateLimiter(rate.Every(time.Hour), 1)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, ip := range []string{"1.1.1.1:80", "2.2.2.2:80", "3.3.3.3:80"} {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		r.RemoteAddr = ip
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("ip %s: status = %d, want 200", ip, w.Code)
		}
	}
}

func TestIPRateLimiter_EvictStale(t *testing.T) {
	rl := newIPRateLimiter(rate.Every(time.Hour), 5)
	rl.ttl = 1 * time.Millisecond // TTL curto para o teste

	// Gera entrada para um IP
	rl.getLimiter("192.168.0.1")

	if len(rl.entries) != 1 {
		t.Fatalf("esperava 1 entrada, got %d", len(rl.entries))
	}

	// Aguarda o TTL expirar e dispara evição manual
	time.Sleep(5 * time.Millisecond)
	rl.evictStale()

	rl.mu.Lock()
	count := len(rl.entries)
	rl.mu.Unlock()

	if count != 0 {
		t.Errorf("esperava 0 entradas apos eviction, got %d", count)
	}
}
