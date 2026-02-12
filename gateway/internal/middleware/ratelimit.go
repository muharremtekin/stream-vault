package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// RateLimiter holds the Redis client and rate limit parameters for the
// token bucket algorithm.
type RateLimiter struct {
	client            *redis.Client
	requestsPerSecond float64
	burst             int
}

// tokenBucketScript is a Lua script executed atomically in Redis.
// It implements a token bucket: tokens are replenished at a fixed rate up to
// a maximum burst size. Each request consumes one token.
var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local bucket = redis.call("HMGET", key, "tokens", "last")
local tokens = tonumber(bucket[1])
local last = tonumber(bucket[2])

if tokens == nil then
    tokens = burst
    last = now
end

local elapsed = math.max(0, now - last)
tokens = math.min(burst, tokens + elapsed * rate)

local allowed = 0
if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

redis.call("HMSET", key, "tokens", tokens, "last", now)
redis.call("EXPIRE", key, math.ceil(burst / rate) + 1)

return allowed
`)

// NewRateLimiter creates a RateLimiter backed by Redis.
func NewRateLimiter(redisURL string, requestsPerSecond float64, burst int) (*RateLimiter, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis url: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}

	return &RateLimiter{
		client:            client,
		requestsPerSecond: requestsPerSecond,
		burst:             burst,
	}, nil
}

// Close shuts down the underlying Redis connection.
func (rl *RateLimiter) Close() error {
	return rl.client.Close()
}

// Middleware returns an http middleware that enforces the token bucket rate
// limit per client IP. When the limit is exceeded it responds with 429.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := extractClientIP(r)
			key := fmt.Sprintf("ratelimit:%s", clientIP)

			ctx := r.Context()
			now := float64(time.Now().UnixMicro()) / 1e6

			allowed, err := tokenBucketScript.Run(ctx, rl.client, []string{key},
				rl.requestsPerSecond,
				rl.burst,
				now,
			).Int()

			if err != nil {
				log.Error().Err(err).Str("client_ip", clientIP).Msg("rate limiter script error")
				// Fail open: allow the request if Redis is unavailable.
				next.ServeHTTP(w, r)
				return
			}

			if allowed == 0 {
				log.Warn().Str("client_ip", clientIP).Msg("rate limit exceeded")
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"rate limit exceeded, try again later"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractClientIP determines the client IP from X-Forwarded-For,
// X-Real-IP, or the connection remote address.
func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first (left-most) IP which is the original client.
		if idx := len(xff); idx > 0 {
			parts := splitFirst(xff, ",")
			return parts
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// splitFirst returns the trimmed string before the first occurrence of sep,
// or the whole trimmed string if sep is not found.
func splitFirst(s, sep string) string {
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			return s[:i]
		}
	}
	return s
}

// NewNoOpRateLimiter returns a middleware that performs no rate limiting.
// Useful when Redis is not available during development.
func NewNoOpRateLimiter() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return next
	}
}
