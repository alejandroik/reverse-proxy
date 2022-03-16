package limiter

import (
	"sync"
	"time"

	"github.com/alejandroik/reverse-proxy/config"
	"github.com/alejandroik/reverse-proxy/logger"

	"golang.org/x/time/rate"
)

type rateConfig struct {
    r   rate.Limit
    b   int
}

type visitor struct {
    RL *rate.Limiter
    conCount int
    lastSeen time.Time
}

// LimiterGroup is a group of rate limiters
type LimiterGroup struct {
    Name string
    interval time.Duration
    data map[string]*visitor
    mu *sync.RWMutex
    RC *rateConfig
}

// InitLimiters returns enabled rate limiter groups
func InitLimiters(c config.Configuration) []*LimiterGroup {
    var limiters []*LimiterGroup

    if c.IP_RATE_ENABLED {
        rc := newRateConfig(c.IP_RATE_LIMIT, c.IP_BURST_LIMIT)
        limiters = append(limiters, newLimiterGroup("IP", rc, c.IP_CLEAN_INTERVAL))
        logger.Info("[IP-Limiter] Started")
    }

    if c.PATH_RATE_ENABLED {
        rc := newRateConfig(c.PATH_RATE_LIMIT, c.PATH_BURST_LIMIT)
        limiters = append(limiters, newLimiterGroup("PATH", rc, c.PATH_CLEAN_INTERVAL))
        logger.Info("[PATH-Limiter] Started")
    }

    for _, lg := range limiters {
        go lg.cleanup()
    }

    return limiters
}

// cleanup removes expired entries from the limiter group
func (lg *LimiterGroup) cleanup() {
	for {
		time.Sleep(lg.interval)
        logger.Infof("[%s-Limiter] Checking for old entries...", lg.Name)

		lg.mu.Lock()
		for k, v := range lg.data {
			if time.Since(v.lastSeen) >= lg.interval {
				delete(lg.data, k)
                logger.Infof("[%s-Limiter] Removed entry for %s", lg.Name, k)
			}
		}
		lg.mu.Unlock()
	}
}

// newRateConfig returns a new rate config
func newRateConfig(r int, b int) *rateConfig {
    return &rateConfig{
        r: rate.Limit(r),
        b: b,
    }
}

// newLimiterGroup returns a new limiter group
func newLimiterGroup(name string, rc *rateConfig, interval int) *LimiterGroup {
    return &LimiterGroup{
        Name: name,
        interval: time.Minute*time.Duration(interval),
        data: make(map[string]*visitor),
        mu:  &sync.RWMutex{},
        RC: rc,
    }
}

// add adds a new rate limiter for a given key
func (lg *LimiterGroup) add(k string) *visitor {
    v := &visitor{
        RL: rate.NewLimiter(lg.RC.r, lg.RC.b),
        conCount: 1,
        lastSeen: time.Now(),
    }

    lg.mu.Lock()
    lg.data[k] = v
    lg.mu.Unlock()

    logger.Infof("[%s-Limiter] Added entry for %s", lg.Name, k)
    
    return v
}

// GetVisitor returns a rate limiter for a given key if it exists, otherwise it creates a new one
func (lg *LimiterGroup) GetVisitor(k string) *visitor {
    lg.mu.RLock()
    v, exists := lg.data[k]
    lg.mu.RUnlock()

    if !exists {
        v = lg.add(k)
    }

    v.lastSeen = time.Now()
    v.conCount++

    return v
}