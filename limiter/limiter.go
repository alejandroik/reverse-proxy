package limiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/alejandroik/reverse-proxy/config"
	"github.com/alejandroik/reverse-proxy/logger"

	"golang.org/x/time/rate"
)

type limiter struct {
    RL *rate.Limiter
    ConCount int
    LastCon time.Time
}

// LimiterGroup is a group of rate limiters
type LimiterGroup struct {
    Name string
    interval time.Duration
    cls map[string]*limiter
    mu *sync.RWMutex
    Limiter *limiter
    RateConfig config.RateConfig
}

type getFunc func() *limiter
type addFunc func(*limiter)

type Limiters map[string]*LimiterGroup

// InitLimiters returns enabled rate limiter groups
func InitLimiters(cfg *config.Config) Limiters {
    var limiters = make(Limiters)

    for _, e := range cfg.Endpoints {
        if e.RateConfig.RateLimit <= 0 && e.RateConfig.ClientRateLimit <= 0 {
            continue
        }
        limiters.AddLimiterGroup(e)
    }

    return limiters
}

func (limiters Limiters) AddLimiterGroup(e config.Endpoint) {
    lg := newLimiterGroup(e.Endpoint, e.RateConfig)
    if e.RateConfig.RateLimit > 0 {
        lg.Limiter = newLimiter(e.RateConfig.RateLimit, e.RateConfig.RateLimit)
    }
    limiters[e.Endpoint] = lg
    go lg.cleanup()
    logger.Info(fmt.Sprintf("[Limiter] Started limiter for %s", lg.Name))
}

// cleanup removes expired entries from the limiter group
func (lg *LimiterGroup) cleanup() {
	for {
		time.Sleep(lg.interval)
        logger.Info(fmt.Sprintf("[Limiter] Checking for old entries in %s", lg.Name))

		lg.mu.Lock()
		for k, v := range lg.cls {
			if time.Since(v.LastCon) >= lg.interval {
				delete(lg.cls, k)
                logger.Info(fmt.Sprintf("[Limiter] Removed entry for %s in %s", k, lg.Name))
			}
		}
		lg.mu.Unlock()
	}
}

func newLimiter(r int, b int) *limiter {
    return &limiter{
        RL: rate.NewLimiter(rate.Limit(r), b),
        ConCount: 0,
        LastCon: time.Now(),
    }
}

// newLimiterGroup returns a new limiter group
func newLimiterGroup(name string, rc config.RateConfig) *LimiterGroup {
    return &LimiterGroup{
        Name: name,
        interval: time.Minute*time.Duration(rc.CleanInterval),
        cls: make(map[string]*limiter),
        mu:  &sync.RWMutex{},
        RateConfig: rc,
    }
}

func (lg *LimiterGroup) GetClientLimiter(k string) *limiter {
    get := func() *limiter {
        lg.mu.RLock()
        l := lg.cls[k]
        lg.mu.RUnlock()

        return l
    }

    add := func(l *limiter) {
        lg.mu.Lock()
        lg.cls[k] = l
        lg.mu.Unlock()

        logger.Info(fmt.Sprintf("[Limiter] Added entry for %s in %s", k, lg.Name))
    }

    return getLimiter(get, add, lg.RateConfig.ClientRateLimit)
}

func getLimiter(get getFunc, add addFunc, r int) *limiter {
    l := get()

    if l == nil {
        l = newLimiter(r, r)
        add(l)
    }

    l.ConCount++
    return l
}

func (lg *LimiterGroup) GetEndpointLimiter() *limiter { 
    get := func() *limiter {
        lg.mu.RLock()
        l := lg.Limiter
        lg.mu.RUnlock()

        return l
    }

    add := func(l *limiter) {
        lg.mu.Lock()
        lg.Limiter = l
        lg.mu.Unlock()

        logger.Info(fmt.Sprintf("[Limiter] Added limiter for %s", lg.Name))
    }

    return getLimiter(get, add, lg.RateConfig.RateLimit)
}

func (lg *LimiterGroup) GetRateConfig() *config.RateConfig {
    lg.mu.RLock()
    rc := lg.RateConfig
    lg.mu.RUnlock()

    return &rc
}