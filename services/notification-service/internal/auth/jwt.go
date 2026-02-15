package auth

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the expected JWT payload structure.
type Claims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

// WSAuthResult holds the extracted identity from a validated JWT.
type WSAuthResult struct {
	UserID string
	Role   string
}

// ValidateWSToken extracts and validates the JWT token from a WebSocket upgrade
// request's query parameter (?token=...). Returns the user identity on success.
func ValidateWSToken(r *http.Request, secret, issuer string) (*WSAuthResult, error) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		return nil, fmt.Errorf("missing token query parameter")
	}

	claims := &Claims{}
	opts := []jwt.ParserOption{
		jwt.WithIssuer(issuer),
		jwt.WithValidMethods([]string{"HS256", "HS384", "HS512"}),
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, opts...)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	if claims.Subject == "" {
		return nil, fmt.Errorf("token missing subject claim")
	}

	return &WSAuthResult{
		UserID: claims.Subject,
		Role:   claims.Role,
	}, nil
}
