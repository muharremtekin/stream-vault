package auth

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testIssuer = "StreamVault.UserService"
	testSecret = "notification-jwt-test-secret-32-bytes-minimum"
)

func signedToken(t *testing.T, secret string) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			Issuer:    testIssuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Role: "Premium",
	})

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("signing token: %v", err)
	}
	return signed
}

func TestValidateWSTokenAcceptsTokenSignedWithConfiguredSecret(t *testing.T) {
	req := httptest.NewRequest("GET", fmt.Sprintf("/ws/notifications?token=%s", signedToken(t, testSecret)), nil)

	result, err := ValidateWSToken(req, testSecret, testIssuer)
	if err != nil {
		t.Fatalf("ValidateWSToken returned error: %v", err)
	}
	if result.UserID != "user-123" || result.Role != "Premium" {
		t.Fatalf("unexpected auth result: %+v", result)
	}
}

func TestValidateWSTokenRejectsTokenSignedWithDifferentSecret(t *testing.T) {
	req := httptest.NewRequest("GET", fmt.Sprintf("/ws/notifications?token=%s", signedToken(t, "different-signing-secret-32-bytes-minimum")), nil)

	if _, err := ValidateWSToken(req, testSecret, testIssuer); err == nil {
		t.Fatal("expected token signed with a different secret to be rejected")
	}
}

func TestValidateWSTokenRejectsTokenSignedWithEmptySecret(t *testing.T) {
	req := httptest.NewRequest("GET", fmt.Sprintf("/ws/notifications?token=%s", signedToken(t, "")), nil)

	if _, err := ValidateWSToken(req, testSecret, testIssuer); err == nil {
		t.Fatal("expected token signed with an empty secret to be rejected")
	}
}

func TestValidateWSTokenRejectsEmptyConfiguredSecret(t *testing.T) {
	req := httptest.NewRequest("GET", fmt.Sprintf("/ws/notifications?token=%s", signedToken(t, testSecret)), nil)

	if _, err := ValidateWSToken(req, " ", testIssuer); err == nil {
		t.Fatal("expected an empty configured secret to be rejected")
	}
}
