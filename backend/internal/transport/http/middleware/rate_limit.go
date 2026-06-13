package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count      int
	windowEnds time.Time
	lastSeen   time.Time
}

func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	var (
		mu       sync.Mutex
		visitors = map[string]*visitor{}
	)

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			for key, v := range visitors {
				if time.Since(v.lastSeen) > 10*time.Minute {
					delete(visitors, key)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		v, ok := visitors[ip]
		if !ok {
			v = &visitor{windowEnds: time.Now().Add(window)}
			visitors[ip] = v
		}
		v.lastSeen = time.Now()
		if time.Now().After(v.windowEnds) {
			v.count = 0
			v.windowEnds = time.Now().Add(window)
		}
		v.count++
		mu.Unlock()

		if v.count > maxRequests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{"code": "rate_limited", "message": "too many requests"},
			})
			return
		}

		c.Next()
	}
}
