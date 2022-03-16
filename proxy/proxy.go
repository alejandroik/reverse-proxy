package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/alejandroik/reverse-proxy/config"
	"github.com/alejandroik/reverse-proxy/logger"
	"github.com/alejandroik/reverse-proxy/utils"
)

// Proxy is a reverse proxy that redirects requests to the remote server
type Proxy struct {
	url *url.URL
	rp *httputil.ReverseProxy
}

// InitProxy initializes a reverse proxy and returns it
func InitProxy(c config.Configuration) *Proxy {
	if strings.TrimSpace(c.REMOTE_URL) == "" {
		logger.Fatal("REMOTE_URL not set")
	}
	url, err := url.Parse(c.REMOTE_URL)
	if err != nil {
		panic(err)
	}
	rp := httputil.NewSingleHostReverseProxy(url)
	return &Proxy{url, rp}
}

// Redirect redirects the request to the remote server
func (p *Proxy) Redirect(w http.ResponseWriter, req *http.Request) {
	ip, err := utils.GetIP(req)
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	req.Host = p.url.Host
	p.rp.ServeHTTP(w, req)
	logger.Infof("%s %s from %s redirected to %s", req.Method, req.URL.Path, ip, p.url.Host)
}