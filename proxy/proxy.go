package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/alejandroik/reverse-proxy/config"
	"github.com/alejandroik/reverse-proxy/logger"
	"github.com/alejandroik/reverse-proxy/utils"
)

type Proxy struct {
	url *url.URL
	rp *httputil.ReverseProxy
}

func InitProxy(c config.Configuration) *Proxy {
	url, err := url.Parse(c.REMOTE_URL)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	return &Proxy{url, proxy}
}

func (p *Proxy) Redirect(w http.ResponseWriter, req *http.Request) {
	ip, err := utils.GetIP(req)
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	req.Host = p.url.Host
	p.rp.ServeHTTP(w, req)
	logger.Infof("%s %s from %s redirected to %s\n", req.Method, req.URL.Path, ip, p.url.Host + req.URL.String())
}