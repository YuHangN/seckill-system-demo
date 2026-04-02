package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// visitor 保存某个 IP 的限流器和最后访问时间。
// lastSeen 用于定期清理长时间不活跃的 IP，防止内存泄漏。
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter 管理所有 IP 的令牌桶限流器。
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rps      rate.Limit // 每秒请求数
	burst    int        // 令牌桶容量（允许的瞬时突发量）
}

// NewRateLimiter 创建限流器，rps=每秒速率，burst=桶容量。
// 例如 NewRateLimiter(10, 10) 表示每秒 10 个请求，最多允许突发 10 个。
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
	// 后台 goroutine 每分钟清理一次超过 3 分钟未访问的 IP。
	go rl.cleanupLoop()
	return rl
}

// getVisitor 返回（或创建）某个 IP 的限流器。
func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rps, rl.burst)
		rl.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupLoop 定期删除超过 3 分钟未访问的 IP 条目。
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	// TODO：考虑优化，因为每分钟扫描所有 IP 会长时间占用锁
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware 返回 Gin 中间件函数，对每个 IP 请求执行限流检查。
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.getVisitor(ip)

		// Allow() 非阻塞：有令牌则消耗一个返回 true，否则返回 false。
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate_limit_exceeded",
			})
			return
		}

		c.Next()
	}
}
