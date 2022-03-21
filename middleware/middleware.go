package middleware

import (
	"fmt"
	"net"
	"net/http"
	"path"

	"github.com/alejandroik/reverse-proxy/limiter"
	"github.com/alejandroik/reverse-proxy/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Middleware struct {
	limiters limiter.Limiters
}

func (m *Middleware) InitMiddleware(limiters limiter.Limiters) {
	m.limiters = limiters

	prometheus.Register(totalRequests)
	prometheus.Register(httpDuration)
}

var (
	totalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Count of all HTTP requests",
		}, []string{"path", "method", "ip"})

	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_response_time_seconds",
			Help: "Duration of HTTP requests.",
		}, []string{"path", "method", "ip"})
)

func Prometheus(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := path.Dir(r.URL.Path)
		method := r.Method
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path, method, ip))

		next.ServeHTTP(w, r)

		totalRequests.WithLabelValues(path, method, ip).Inc()
		timer.ObserveDuration()
	})
}

// Limit checks the request rate and returns a handler that returns 429 if the rate is exceeded
func (m *Middleware) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if lg, ok := m.limiters["/"]; ok {
			allowed := getLimiters(lg, w, req)
			if !allowed {
				return
			}
		}

		lg, ok := m.limiters[path.Dir(req.URL.Path)]
		if !ok {
			next.ServeHTTP(w, req)
			return
		}

		allowed := getLimiters(lg, w, req)
		if !allowed {
			return
		}

		next.ServeHTTP(w, req)
	})
}

func getLimiters(lg *limiter.LimiterGroup, w http.ResponseWriter, req *http.Request) bool {
	if lg.GetRateConfig().ClientRateLimit > 0 {
		ip, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		if !lg.GetClientLimiter(ip).RL.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			logger.Info(fmt.Sprintf("[Limiter] Denied request to %s for %s", lg.Name, ip))
			return false
		}
	}

	if lg.GetRateConfig().RateLimit > 0 && !lg.GetEndpointLimiter().RL.Allow() {
		logger.Info(fmt.Sprintf("[Limiter] Rate limit exceeded for %s", lg.Name))
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return false
	}

	return true
}
