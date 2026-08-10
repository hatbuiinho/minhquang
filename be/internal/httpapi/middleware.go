package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const requestIDHeader = "X-Request-ID"

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	written, err := r.ResponseWriter.Write(body)
	r.bytes += written
	return written, err
}

func withRequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = newRequestID()
		}
		w.Header().Set(requestIDHeader, requestID)

		if shouldSkipRequestLog(r) {
			next.ServeHTTP(w, r)
			return
		}

		startedAt := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}

		log.Printf(
			"http_request method=%s path=%s status=%d duration_ms=%d bytes=%d remote_ip=%s request_id=%s user_agent=%q",
			r.Method,
			r.URL.RequestURI(),
			status,
			time.Since(startedAt).Milliseconds(),
			recorder.bytes,
			remoteIP(r),
			requestID,
			r.UserAgent(),
		)
	})
}

func shouldSkipRequestLog(r *http.Request) bool {
	return r.Method == http.MethodOptions || r.URL.Path == "/healthz"
}

func newRequestID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(bytes[:])
}

func remoteIP(r *http.Request) string {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := r.Header.Get(header)
		if value == "" {
			continue
		}
		value = strings.TrimSpace(strings.Split(value, ",")[0])
		host, _, err := net.SplitHostPort(value)
		if err == nil {
			return host
		}
		return value
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
