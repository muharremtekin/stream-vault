package middleware

import (
	"net/url"
	"strings"
	"testing"
)

func TestRedactSensitiveQuery(t *testing.T) {
	raw := "window=week&limit=10&token=do-not-log&REFRESH_TOKEN=also-secret"
	redacted := redactSensitiveQuery(raw)

	if strings.Contains(redacted, "do-not-log") || strings.Contains(redacted, "also-secret") {
		t.Fatalf("sensitive query value was not redacted: %q", redacted)
	}

	values, err := url.ParseQuery(redacted)
	if err != nil {
		t.Fatalf("redacted query is invalid: %v", err)
	}
	if values.Get("window") != "week" || values.Get("limit") != "10" {
		t.Fatalf("non-sensitive query values changed: %q", redacted)
	}
	if values.Get("token") != "[REDACTED]" || values.Get("REFRESH_TOKEN") != "[REDACTED]" {
		t.Fatalf("sensitive query values were not replaced: %q", redacted)
	}
}

func TestRedactSensitiveQuery_InvalidEncoding(t *testing.T) {
	if got := redactSensitiveQuery("token=%zz"); got != "[invalid query]" {
		t.Fatalf("expected invalid query marker, got %q", got)
	}
}
