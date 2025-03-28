package http

import (
	"context"
	"net/http"

	"U_PAY/db/rdb"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
)

// RateLimitMiddleware 创建一个限流中间件
func RateLimitMiddleware() gin.HandlerFunc {
	// 创建一个新的限流器
	limiter := redis_rate.NewLimiter(rdb.RDB)

	return func(c *gin.Context) {
		// 获取客户端真实IP
		realIP := c.GetHeader("X-Real-IP")
		if realIP == "" {
			realIP = c.GetHeader("X-Forwarded-For")
			if realIP == "" {
				realIP = c.ClientIP()
			}
		}
		// log.Logger.Info("当前用户IP地址: ", zap.String("realIP", realIP))

		key := "rate_limit:" + realIP

		// 限制每个IP每分钟最多100个请求
		res, err := limiter.Allow(context.Background(), key, redis_rate.PerMinute(100))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "限流服务异常",
			})
			return
		}

		// 如果超过限制，返回429状态码
		if res.Allowed == 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "请求过于频繁，请稍后再试",
				"retry_after": res.RetryAfter.Seconds(),
			})
			return
		}

		c.Next()
	}
}
