package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alejandroik/reverse-proxy/config"
	"github.com/alejandroik/reverse-proxy/limiter"
)

func TestLimiterMiddleware(t *testing.T) {
	// Given
	var cfg *config.Config = &config.Config{}
	var ok, ko int

	r := 2
	cfg.Limiters = append(cfg.Limiters, config.Limiter{
		Endpoint: "/",
		RateConfig: config.RateConfig{
			RateLimit:     r,
			CleanInterval: 10,
		},
	})

	l := limiter.InitLimiters(cfg)
	m := Middleware{}
	m.InitMiddleware(l)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// When
	handler := m.Limit(next)
	total := 3
	for i := 0; i < total; i++ {
		req, _ := http.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		if rr.Result().StatusCode == 200 {
			ok++
			continue
		}
		if rr.Result().StatusCode == 429 || rr.Result().StatusCode == 503 {
			ko++
			continue
		}
	}

	// Then
	if ok != 2 && ko != 1 {
		t.Errorf("OK requests %d exceded rate limit %d", ok, r)
	}
}
