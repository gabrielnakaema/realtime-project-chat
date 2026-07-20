package httpmw

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/platform/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func Mount(r chi.Router, corsOrigins []string, log *slog.Logger) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   corsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(RequestLogging(log))
	r.Use(middleware.Recoverer)
}

func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}
}

func RequestLogging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			l := log
			if requestID := middleware.GetReqID(r.Context()); requestID != "" {
				l = l.With("request_id", requestID)
			}
			r = r.WithContext(logger.WithLogger(r.Context(), l))

			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			duration := time.Since(start)

			logLevel := slog.LevelInfo
			if ww.statusCode >= 400 && ww.statusCode < 500 {
				logLevel = slog.LevelWarn
			} else if ww.statusCode >= 500 {
				logLevel = slog.LevelError
			}

			ip := r.RemoteAddr
			if ip == "" {
				ip = r.Header.Get("X-Forwarded-For")
			}

			l.Log(r.Context(), logLevel, "http_request",
				"method", r.Method,
				"url", r.URL.String(),
				"remote_addr", ip,
				"status", ww.statusCode,
				"duration_ms", duration.Milliseconds(),
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	hijacked   bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.hijacked {
		return
	}
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
	}
	rw.hijacked = true
	return hijacker.Hijack()
}
