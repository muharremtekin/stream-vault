package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// publicPrefixes lists URL path prefixes that bypass JWT authentication.
var publicPrefixes = []string{
	"/api/auth/",
	"/api/catalog/",
	"/api/search/trending",
	"/api/plans/",
	"/health",
	"/ws/",
}

// Claims represents the expected JWT payload structure.
type Claims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

// Auth returns middleware that validates JWT bearer tokens. Requests to public
// routes are forwarded without authentication. On success the middleware injects
// X-User-Id and X-User-Role headers so upstream services can identify the caller.
func Auth(secret string, issuer string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicRoute(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			tokenString, ok := extractBearerToken(r)
			if !ok {
				WriteErrorResponse(w, http.StatusUnauthorized, "missing or malformed authorization header")
				return
			}

			claims, err := parseToken(tokenString, secret, issuer)
			if err != nil {
				log.Warn().Err(err).Str("path", r.URL.Path).Msg("jwt validation failed")
				WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			// Propagate identity to upstream services via headers.
			r.Header.Set("X-User-Id", claims.Subject)
			r.Header.Set("X-User-Role", claims.Role)

			next.ServeHTTP(w, r)
		})
	}
}

// isPublicRoute checks whether the given path matches any public prefix.
func isPublicRoute(path string) bool {
	for _, prefix := range publicPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	// Exact matches for paths without trailing slash.
	if path == "/health" || path == "/api/plans" {
		return true
	}
	return false
}

// extractBearerToken pulls the token string from the Authorization header.
func extractBearerToken(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", false
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", false
	}
	return parts[1], true
}

// parseToken validates and parses the JWT token string against the provided
// secret and issuer.
func parseToken(tokenString, secret, issuer string) (*Claims, error) {
	claims := &Claims{}

	opts := []jwt.ParserOption{
		jwt.WithIssuer(issuer),
		jwt.WithValidMethods([]string{"HS256", "HS384", "HS512"}),
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, opts...)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenUnverifiable
	}

	return claims, nil
}
