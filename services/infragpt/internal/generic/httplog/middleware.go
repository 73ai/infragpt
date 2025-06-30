package httplog

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.body != nil && len(b) < 1024 { // Only capture small responses
		rw.body.Write(b)
	}
	return rw.ResponseWriter.Write(b)
}

// Middleware returns an HTTP middleware that logs requests and responses with colorful output
// If enabled is false, it returns the handler unchanged
func Middleware(enabled bool) func(http.Handler) http.Handler {
	if !enabled {
		return func(h http.Handler) http.Handler {
			return h
		}
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			var requestBody []byte
			if r.Body != nil {
				requestBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}

			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           &bytes.Buffer{},
			}
			methodColor := getMethodColor(r.Method)
			slog.Info(fmt.Sprintf("%s%s%s %s%s%s started",
				colorBold, methodColor, r.Method,
				colorCyan, r.URL.Path, colorReset),
				"remote_addr", r.RemoteAddr,
				"content_length", r.ContentLength,
				"auth_header_present", r.Header.Get("Authorization") != "",
				"request_body", string(requestBody),
			)

			h.ServeHTTP(rw, r)
			duration := time.Since(start)
			statusColor := getStatusColor(rw.statusCode)
			slog.Info(fmt.Sprintf("%s%s%s %s%s%s completed %s%d%s in %s%dms%s",
				colorBold, methodColor, r.Method,
				colorCyan, r.URL.Path, colorReset,
				statusColor, rw.statusCode, colorReset,
				colorYellow, duration.Milliseconds(), colorReset),
				"response_body", rw.body.String(),
			)
		})
	}
}

func getMethodColor(method string) string {
	switch method {
	case "GET":
		return colorGreen
	case "POST":
		return colorBlue
	case "PUT":
		return colorYellow
	case "DELETE":
		return colorRed
	case "PATCH":
		return colorPurple
	default:
		return colorWhite
	}
}

func getStatusColor(status int) string {
	switch {
	case status >= 200 && status < 300:
		return colorGreen
	case status >= 300 && status < 400:
		return colorYellow
	case status >= 400 && status < 500:
		return colorRed
	case status >= 500:
		return colorRed + colorBold
	default:
		return colorWhite
	}
}
