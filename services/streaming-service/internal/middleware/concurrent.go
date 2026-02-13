package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/streaming-service/internal/handler"
)

var concurrentCheckScript = redis.NewScript(`
local key = KEYS[1]
local session_id = ARGV[1]
local now = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])
local max_streams = tonumber(ARGV[4])

redis.call("ZREMRANGEBYSCORE", key, "-inf", now - ttl)

local existing = redis.call("ZSCORE", key, session_id)
if existing then
    redis.call("ZADD", key, now, session_id)
    redis.call("EXPIRE", key, ttl + 60)
    return 1
end

local count = redis.call("ZCARD", key)
if count >= max_streams then
    return 0
end

redis.call("ZADD", key, now, session_id)
redis.call("EXPIRE", key, ttl + 60)
return 1
`)

func ConcurrentStreams(redisClient *redis.Client, concurrentTTL time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Header.Get("X-User-Id")
			if userID == "" {
				handler.WriteErrorResponse(w, http.StatusUnauthorized, "authentication required")
				return
			}

			role := r.Header.Get("X-User-Role")
			maxStreams := tierToMaxStreams(role)

			sessionID := r.Header.Get("X-Session-Id")
			if sessionID == "" {
				sessionID = userID + ":" + clientIP(r)
			}

			key := "concurrent:" + userID
			now := time.Now().Unix()
			ttlSec := int64(concurrentTTL.Seconds())

			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			result, err := concurrentCheckScript.Run(ctx, redisClient, []string{key},
				sessionID, now, ttlSec, maxStreams).Int64()
			if err != nil {
				log.Warn().Err(err).Msg("concurrent stream check failed, allowing request")
				next.ServeHTTP(w, r)
				return
			}

			if result == 0 {
				w.Header().Set("X-Max-Streams", strconv.Itoa(maxStreams))
				handler.WriteErrorResponse(w, http.StatusTooManyRequests, "concurrent stream limit exceeded")
				return
			}

			rec := newResponseRecorder(w)
			next.ServeHTTP(rec, r)

			// Remove phantom session if handler returned non-2xx
			if rec.statusCode < 200 || rec.statusCode >= 300 {
				cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cleanupCancel()
				if err := redisClient.ZRem(cleanupCtx, key, sessionID).Err(); err != nil {
					log.Warn().Err(err).Str("user_id", userID).Str("session_id", sessionID).
						Msg("failed to remove phantom concurrent session")
				}
			}
		})
	}
}

func tierToMaxStreams(role string) int {
	switch strings.ToLower(role) {
	case "admin", "premium", "subscription_tier_premium":
		return 4
	case "standard", "subscription_tier_standard":
		return 2
	default:
		return 1
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}
