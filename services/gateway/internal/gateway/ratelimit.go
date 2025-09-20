package gateway

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// RateLimitMiddleware handles rate limiting
type RateLimitMiddleware struct {
	redisClient *redis.Client
	limiters    map[string]*rate.Limiter // In-memory limiters for fallback
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(config *RateLimitConfig, redisClient *redis.Client) (*RateLimitMiddleware, error) {
	return &RateLimitMiddleware{
		redisClient: redisClient,
		limiters:    make(map[string]*rate.Limiter),
	}, nil
}

// Handler returns the rate limit middleware handler
func (r *RateLimitMiddleware) Handler(config *RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.Enabled {
			c.Next()
			return
		}
		
		// Generate rate limit key
		key := r.generateKey(c, config)
		
		// Get rate limit for this key
		limit := r.getLimitForKey(c, config)
		
		// Check rate limit using Redis
		allowed, remaining, resetTime, err := r.checkRateLimit(key, limit.RequestsPerMinute, limit.BurstSize)
		if err != nil {
			// Fallback to in-memory rate limiting
			allowed = r.checkInMemoryRateLimit(key, limit.RequestsPerMinute, limit.BurstSize)
			if !allowed {
				r.sendRateLimitResponse(c, 0, 0, time.Now().Add(time.Minute))
				return
			}
		} else {
			// Set rate limit headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(limit.RequestsPerMinute))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
			
			if !allowed {
				r.sendRateLimitResponse(c, limit.RequestsPerMinute, remaining, resetTime)
				return
			}
		}
		
		c.Next()
	}
}

// generateKey generates a rate limit key based on the strategy
func (r *RateLimitMiddleware) generateKey(c *gin.Context, config *RateLimitConfig) string {
	switch config.KeyStrategy {
	case "ip":
		return "ratelimit:ip:" + c.ClientIP()
	case "user":
		if userID, exists := c.Get("user_id"); exists {
			return "ratelimit:user:" + userID.(string)
		}
		// Fallback to IP if no user
		return "ratelimit:ip:" + c.ClientIP()
	case "tenant":
		if tenantID, exists := c.Get("tenant_id"); exists {
			return "ratelimit:tenant:" + tenantID.(string)
		}
		// Fallback to IP if no tenant
		return "ratelimit:ip:" + c.ClientIP()
	case "custom":
		if config.CustomKey != "" {
			// Extract custom key from headers, query params, etc.
			if customValue := c.GetHeader(config.CustomKey); customValue != "" {
				return "ratelimit:custom:" + customValue
			}
			if customValue := c.Query(config.CustomKey); customValue != "" {
				return "ratelimit:custom:" + customValue
			}
		}
		// Fallback to IP
		return "ratelimit:ip:" + c.ClientIP()
	default:
		return "ratelimit:ip:" + c.ClientIP()
	}
}

// getLimitForKey gets the appropriate rate limit for the request
func (r *RateLimitMiddleware) getLimitForKey(c *gin.Context, config *RateLimitConfig) RateLimit {
	// Check if user has specific limits based on roles/scopes
	if userRoles, exists := c.Get("user_roles"); exists {
		roles := userRoles.([]string)
		for _, role := range roles {
			if limit, ok := config.Limits[role]; ok {
				return limit
			}
		}
	}
	
	// Check tenant-specific limits
	if tenantID, exists := c.Get("tenant_id"); exists {
		if limit, ok := config.Limits[tenantID.(string)]; ok {
			return limit
		}
	}
	
	// Default limit
	return RateLimit{
		RequestsPerMinute: config.RequestsPerMinute,
		BurstSize:         config.BurstSize,
	}
}

// checkRateLimit checks rate limit using Redis sliding window
func (r *RateLimitMiddleware) checkRateLimit(key string, requestsPerMinute, burstSize int) (bool, int, time.Time, error) {
	ctx := context.Background()
	now := time.Now()
	window := time.Minute
	
	// Use Redis sorted sets for sliding window rate limiting
	pipe := r.redisClient.Pipeline()
	
	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(now.Add(-window).UnixNano(), 10))
	
	// Count current requests in window
	pipe.ZCard(ctx, key)
	
	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})
	
	// Set expiration
	pipe.Expire(ctx, key, window+time.Second)
	
	results, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, err
	}
	
	// Get current count (before adding new request)
	currentCount := results[1].(*redis.IntCmd).Val()
	
	// Check if within limit
	allowed := currentCount < int64(requestsPerMinute)
	remaining := int(int64(requestsPerMinute) - currentCount - 1)
	if remaining < 0 {
		remaining = 0
	}
	
	resetTime := now.Add(window)
	
	return allowed, remaining, resetTime, nil
}

// checkInMemoryRateLimit fallback rate limiting using in-memory limiters
func (r *RateLimitMiddleware) checkInMemoryRateLimit(key string, requestsPerMinute, burstSize int) bool {
	// Get or create limiter for this key
	limiter, exists := r.limiters[key]
	if !exists {
		// Create new limiter
		limit := rate.Limit(float64(requestsPerMinute) / 60.0) // Convert per-minute to per-second
		limiter = rate.NewLimiter(limit, burstSize)
		r.limiters[key] = limiter
	}
	
	return limiter.Allow()
}

// sendRateLimitResponse sends rate limit exceeded response
func (r *RateLimitMiddleware) sendRateLimitResponse(c *gin.Context, limit, remaining int, resetTime time.Time) {
	c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
	c.Header("Retry-After", strconv.FormatInt(int64(time.Until(resetTime).Seconds()), 10))
	
	c.JSON(http.StatusTooManyRequests, gin.H{
		"error":   "Rate limit exceeded",
		"message": "Too many requests. Please try again later.",
		"retry_after": int64(time.Until(resetTime).Seconds()),
	})
	c.Abort()
}

// CleanupLimiters periodically cleans up unused in-memory limiters
func (r *RateLimitMiddleware) CleanupLimiters() {
	// This should be called periodically to prevent memory leaks
	// In a production system, you might want to use a more sophisticated cleanup strategy
	r.limiters = make(map[string]*rate.Limiter)
}
