package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/alejandroik/reverse-proxy/config"
	"github.com/alejandroik/reverse-proxy/logger"
)

// Proxy is a reverse proxy that redirects requests to the remote server
type Proxy struct {
	url *url.URL
	rp *httputil.ReverseProxy
}

// InitProxy initializes a reverse proxy and returns it
func InitProxy(cfg *config.Config) *Proxy {
	if strings.TrimSpace(cfg.Server.RemoteHost) == "" {
		logger.Fatal("remote_host is not set")
	}
	url, err := url.Parse(cfg.Server.RemoteHost)
	if err != nil {
		panic(err)
	}
	rp := httputil.NewSingleHostReverseProxy(url)
	return &Proxy{url, rp}
}

// Redirect redirects the request to the remote server
func (p *Proxy) Redirect(w http.ResponseWriter, req *http.Request) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	req.Host = p.url.Host
	p.rp.ServeHTTP(w, req)
	logger.Info(fmt.Sprintf("%s %s from %s redirected to %s", req.Method, req.URL.Path, ip, p.url.Host))
}