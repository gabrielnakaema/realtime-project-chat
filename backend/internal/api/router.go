package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (a *Api) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   a.config.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Use(a.addLoggerMiddleware)
	r.Use(a.slogMiddleware)

	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(30 * time.Second))

	r.Use(a.handlers.AuthMiddleware.IdentifyUser)

	r.Route("/users", func(r chi.Router) {
		r.Post("/", a.handlers.User.Create)
		r.Get("/me", a.handlers.User.GetMe)
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", a.handlers.User.Login)
		r.Post("/refresh-token", a.handlers.User.RefreshToken)
	})

	r.Route("/projects", func(r chi.Router) {
		r.Use(a.handlers.AuthMiddleware.ProtectRoutes)
		r.Post("/", a.handlers.Project.Create)
		r.Get("/", a.handlers.Project.List)
		r.Get("/{id}", a.handlers.Project.Get)
		r.Put("/{id}", a.handlers.Project.Update)
		r.Post("/{id}/members", a.handlers.Project.CreateMember)
	})

	r.Route("/tasks", func(r chi.Router) {
		r.Use(a.handlers.AuthMiddleware.ProtectRoutes)
		r.Get("/", a.handlers.Task.List)
		r.Post("/", a.handlers.Task.Create)
		r.Get("/{id}", a.handlers.Task.Get)
		r.Put("/{id}", a.handlers.Task.Update)
	})

	return r
}

func (a *Api) addLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := logger.WithLogger(r.Context(), a.logger)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (a *Api) slogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		l := logger.FromContext(r.Context())
		requestID := middleware.GetReqID(r.Context())
		if requestID != "" {
			l = l.With("request_id", requestID)
			r = r.WithContext(logger.WithLogger(r.Context(), l))
		}

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
