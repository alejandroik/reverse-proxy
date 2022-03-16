package main

import (
	"net/http"

	"github.com/alejandroik/reverse-proxy/config"
	"github.com/alejandroik/reverse-proxy/limiter"
	"github.com/alejandroik/reverse-proxy/logger"
	"github.com/alejandroik/reverse-proxy/middleware"
	"github.com/alejandroik/reverse-proxy/proxy"
)

func main() {
	var c config.Configuration = config.GetConfig()

	mux := http.NewServeMux()

	p := proxy.InitProxy(c)
	mux.HandleFunc("/", p.Redirect)

	l := limiter.InitLimiters(c)
	logger.Infof("Listening on %s", c.SERVER_PORT)
	logger.Fatal(http.ListenAndServe(":"+c.SERVER_PORT, middleware.Limit(mux, l)))
}