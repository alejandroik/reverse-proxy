package main

import (
	"fmt"
	"net"
	"net/http"
	"path"

	"github.com/alejandroik/reverse-proxy/config"
	"github.com/alejandroik/reverse-proxy/limiter"
	"github.com/alejandroik/reverse-proxy/logger"
	"github.com/alejandroik/reverse-proxy/middleware"
	"github.com/alejandroik/reverse-proxy/proxy"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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

func prometheusMiddleware(next http.Handler) http.Handler {
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

func init() {
	prometheus.Register(totalRequests)
	prometheus.Register(httpDuration)
}

func main() {
	cfg := config.GetConfig()

	r := mux.NewRouter()

	l := limiter.InitLimiters(cfg)
	p := proxy.InitProxy(cfg)

	m := middleware.Middleware{}
	m.InitMiddleware(l, cfg)

	r.Use(prometheusMiddleware)
	r.Use(m.Limit)

	r.Path("/metrics").Handler(promhttp.Handler())
	r.PathPrefix("/").HandlerFunc(p.Redirect)

	logger.Info(fmt.Sprintf("Listening on %s", cfg.Server.Port))
	err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Server.Port), r)
	logger.Fatal(err.Error())
}