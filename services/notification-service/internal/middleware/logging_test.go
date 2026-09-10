package middleware

import (
	"bufio"
	"net"
	"net/http"
	"testing"
)

type hijackableResponseWriter struct {
	header   http.Header
	hijacked bool
}

func (w *hijackableResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *hijackableResponseWriter) Write(body []byte) (int, error) { return len(body), nil }
func (w *hijackableResponseWriter) WriteHeader(int)                {}

func (w *hijackableResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	return nil, nil, nil
}

func TestResponseRecorderPreservesHijacker(t *testing.T) {
	underlying := &hijackableResponseWriter{}
	recorder := newResponseRecorder(underlying)

	if _, ok := any(recorder).(http.Hijacker); !ok {
		t.Fatal("response recorder must implement http.Hijacker")
	}
	if _, _, err := recorder.Hijack(); err != nil {
		t.Fatalf("unexpected hijack error: %v", err)
	}
	if !underlying.hijacked {
		t.Fatal("hijack was not delegated to the underlying writer")
	}
}
