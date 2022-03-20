package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alejandroik/reverse-proxy/config"
	"github.com/alejandroik/reverse-proxy/limiter"
)

func TestLimiterMiddleware(t *testing.T) {
	t.Run("Should return a handler that returns 429 if the rate is exceeded", func(t *testing.T) {
		// Given
		cfg := config.GetConfig()
		l := limiter.InitLimiters(cfg)
		m := Middleware{}
		m.InitMiddleware(l, cfg)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// When
		handler := m.Limit(next)
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// Then
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})
}