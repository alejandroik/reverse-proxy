package middleware

import (
	"net/http"

	"github.com/alejandroik/reverse-proxy/limiter"
	"github.com/alejandroik/reverse-proxy/logger"
	"github.com/alejandroik/reverse-proxy/utils"
)

func Limit(next http.Handler, limiters []*limiter.LimiterGroup) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        for _, lg := range limiters {
            p, err := utils.GetParameter(lg.Name, req)
            if err != nil {
                logger.Info(err.Error())
                http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
            }

            v := lg.GetVisitor(p)
            if !v.RL.Allow(){
                http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
                logger.Infof("[%s-Limiter] Denied request for %s", lg.Name, p)
                return
            }
        }

        next.ServeHTTP(w, req)
    })
}