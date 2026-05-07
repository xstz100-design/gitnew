package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────── IP 全局速率限制 ───────────────────────────

type ipBucket struct {
	count    int
	windowAt time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*ipBucket
	maxRPS  int           // 每秒最大请求数
	window  time.Duration // 时间窗口
}

func NewRateLimiter(maxRPS int) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*ipBucket),
		maxRPS:  maxRPS,
		window:  time.Second,
	}
	// 定时清理过期桶
	go func() {
		for {
			time.Sleep(60 * time.Second)
			rl.mu.Lock()
			now := time.Now()
			for ip, b := range rl.buckets {
				if now.Sub(b.windowAt) > 60*time.Second {
					delete(rl.buckets, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[ip]
	if !ok || now.Sub(b.windowAt) >= rl.window {
		rl.buckets[ip] = &ipBucket{count: 1, windowAt: now}
		return true
	}
	b.count++
	return b.count <= rl.maxRPS
}

var globalLimiter = NewRateLimiter(30)

// RateLimit 全局 IP 速率限制中间件（30 req/s）
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !globalLimiter.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"detail": "请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}

// ─────────────────────────── 登录防暴力破解 ───────────────────────────

type loginAttempt struct {
	failures int
	lockedAt *time.Time
}

type LoginProtector struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
	maxFails int
	lockSecs int
}

var LoginGuard = &LoginProtector{
	attempts: make(map[string]*loginAttempt),
	maxFails: 5,
	lockSecs: 300,
}

func (lp *LoginProtector) Check(ip, username string) error {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	key := ip + "|" + username
	a, ok := lp.attempts[key]
	if !ok {
		return nil
	}
	if a.lockedAt != nil {
		if time.Since(*a.lockedAt) < time.Duration(lp.lockSecs)*time.Second {
			return fmt.Errorf("账号已被临时锁定，请 %d 分钟后再试", lp.lockSecs/60)
		}
		// 锁定已过期，重置
		delete(lp.attempts, key)
	}
	return nil
}

func (lp *LoginProtector) RecordFailure(ip, username string) {
	lp.mu.Lock()
	defer lp.mu.Unlock()

	key := ip + "|" + username
	a, ok := lp.attempts[key]
	if !ok {
		a = &loginAttempt{}
		lp.attempts[key] = a
	}
	a.failures++
	if a.failures >= lp.maxFails {
		now := time.Now()
		a.lockedAt = &now
	}
}

func (lp *LoginProtector) Reset(ip, username string) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	delete(lp.attempts, ip+"|"+username)
}
