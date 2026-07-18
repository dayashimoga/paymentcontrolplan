// Package middleware provides HTTP middleware for the PCP API.
package middleware

import (
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

// Recovery returns middleware that recovers from panics, logs the stack trace,
// and returns a 500 Internal Server Error response.
func Recovery(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					logger.Error("panic recovered",
						zap.Any("error", rvr),
						zap.String("stack", string(debug.Stack())),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
					)
					http.Error(w, `{"error":"internal_error","message":"internal server error","code":500}`, http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
