package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type contextKey string

const requestIDKey contextKey = "request_id"

func withRequestID(r *http.Request, requestID string) *http.Request {
	ctx := context.WithValue(r.Context(), requestIDKey, requestID)
	return r.WithContext(ctx)
}

func RequestID(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "request-id-unavailable"
	}

	return hex.EncodeToString(b[:])
}
