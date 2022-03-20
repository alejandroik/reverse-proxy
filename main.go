package main

import (
	"fmt"
	"net/http"

	"github.com/alejandroik/reverse-proxy/config"
	"github.com/alejandroik/reverse-proxy/limiter"
	"github.com/alejandroik/reverse-proxy/logger"
	"github.com/alejandroik/reverse-proxy/middleware"
	"github.com/alejandroik/reverse-proxy/proxy"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.GetConfig()

	r := mux.NewRouter()

	l := limiter.InitLimiters(cfg)
	p := proxy.InitProxy(cfg)

	m := middleware.Middleware{}
	m.InitMiddleware(l, cfg)

	r.Use(middleware.Prometheus, m.Limit)

	r.Path("/metrics").Handler(promhttp.Handler())
	r.PathPrefix("/").HandlerFunc(p.Redirect)

	logger.Info(fmt.Sprintf("Listening on %s", cfg.Server.Port))
	err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Server.Port), r)
	logger.Fatal(err.Error())
}