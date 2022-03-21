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
	RL      *rate.Limiter
	LastCon time.Time
}

// LimiterGroup is a group of rate limiters
type LimiterGroup struct {
	Name       string
	interval   time.Duration
	cls        map[string]*limiter
	mu         *sync.RWMutex
	Limiter    *limiter
	RateConfig config.RateConfig
}

type getFunc func() *limiter
type addFunc func(*limiter)

type Limiters map[string]*LimiterGroup

var default_clean_interval = time.Minute * 10

// InitLimiters returns enabled rate limiter groups
func InitLimiters(cfg *config.Config) Limiters {
	var limiters = make(Limiters)

	for _, cl := range cfg.Limiters {
		if cl.RateConfig.RateLimit <= 0 && cl.RateConfig.ClientRateLimit <= 0 {
			continue
		}
		limiters.AddLimiterGroup(cl)
	}

	return limiters
}

func (limiters Limiters) AddLimiterGroup(cl config.Limiter) {
	lg := newLimiterGroup(cl.Endpoint, cl.RateConfig)
	if cl.RateConfig.RateLimit > 0 {
		lg.Limiter = newLimiter(cl.RateConfig.RateLimit, cl.RateConfig.RateLimit)
	}
	limiters[cl.Endpoint] = lg
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
		RL:      rate.NewLimiter(rate.Limit(r), b),
		LastCon: time.Now(),
	}
}

// newLimiterGroup returns a new limiter group
func newLimiterGroup(name string, rc config.RateConfig) *LimiterGroup {
	var interval time.Duration
	if rc.CleanInterval == 0 {
		interval = default_clean_interval
	} else {
		interval = time.Minute * time.Duration(rc.CleanInterval)
	}
	return &LimiterGroup{
		Name:       name,
		interval:   interval,
		cls:        make(map[string]*limiter),
		mu:         &sync.RWMutex{},
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

	l.LastCon = time.Now()
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
