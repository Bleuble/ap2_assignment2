package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimiterMiddleware(redisClient *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("ratelimit:%s", clientIP)

		ctx := context.Background()

		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {

			c.Next()
			return
		}

		if count == 1 {

			redisClient.Expire(ctx, key, window)
		}

		if count > int64(limit) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": fmt.Sprintf("Rate limit exceeded. Maximum %d requests per %v allowed.", limit, window),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
