package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	body        *bytes.Buffer
}

// Ensure responseWriter implements http.Flusher when the underlying writer does.
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		body:           bytes.NewBuffer(nil),
	}
}

func (rw *responseWriter) Status() int {
	if !rw.wroteHeader {
		return http.StatusOK
	}
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	rw.body.Write(b) // Capture response body
	return rw.ResponseWriter.Write(b)
}

// AccessLog creates a middleware that logs HTTP requests and responses
func AccessLog(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path
			method := r.Method

			// Bypass logging for SSE endpoints to allow proper streaming
			if path == "/sse" || path == "/mcp/sse" {
				next.ServeHTTP(w, r)
				return
			}

			// Read request body
			var reqBody []byte
			if r.Body != nil {
				reqBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
			}

			rw := newResponseWriter(w)
			// Ensure the writer supports flushing for SSE streams.
			if _, ok := rw.ResponseWriter.(http.Flusher); !ok {
				// If the underlying writer does not support Flusher, we cannot stream SSE properly.
				// Proceed without wrapping to avoid breaking SSE.
				next.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			// Log details
			fields := []zap.Field{
				zap.String("method", method),
				zap.String("path", path),
				zap.Int("status", rw.Status()),
				zap.Duration("duration", duration),
				zap.String("ip", r.RemoteAddr),
			}

			// Add bodies for non-SSE requests and if content is text/json
			if path != "/sse" {
				if len(reqBody) > 0 {
					fields = append(fields, zap.String("request_body", string(reqBody)))
				}
				if rw.body.Len() > 0 {
					fields = append(fields, zap.String("response_body", rw.body.String()))
				}
			}

			logger.Info("HTTP Access", fields...)
		})
	}
}
