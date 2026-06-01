package api

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (a *Api) Router() http.Handler {
	r := chi.NewRouter()

	a.logger.Info("CORS Origins", "origins", a.config.CORSOrigins)

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

	r.Handle("/mcp", a.handlers.MCP)

	r.Group(func(r chi.Router) {
		r.Use(a.handlers.AuthMiddleware.IdentifyUser)

		r.Route("/users", func(r chi.Router) {
			r.Post("/", a.handlers.User.Create)
			r.Get("/me", a.handlers.User.GetMe)
			r.Get("/", a.handlers.User.ListUsers)
			r.Route("/me/mcp-api-keys", func(r chi.Router) {
				r.Use(a.handlers.AuthMiddleware.ProtectRoutes)
				r.Post("/", a.handlers.MCPAPIKey.Create)
				r.Get("/", a.handlers.MCPAPIKey.List)
				r.Delete("/{id}", a.handlers.MCPAPIKey.Revoke)
			})
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", a.handlers.User.Login)
			r.Post("/refresh-token", a.handlers.User.RefreshToken)
			r.Post("/logout", a.handlers.User.Logout)
		})

		r.Route("/projects", func(r chi.Router) {
			r.Use(a.handlers.AuthMiddleware.ProtectRoutes)
			r.Post("/", a.handlers.Project.Create)
			r.Get("/", a.handlers.Project.List)
			r.Get("/activities", a.handlers.Project.ListUsersProjectActivities)
			r.Get("/{id}", a.handlers.Project.Get)
			r.Put("/{id}", a.handlers.Project.Update)
			r.Get("/{id}/activities", a.handlers.Project.ListActivitiesByProject)
			r.Post("/{id}/members", a.handlers.Project.CreateMember)
			r.Get("/{id}/members", a.handlers.Project.ListMembersByProjectId)
			r.Delete("/{id}/members/{member_id}", a.handlers.Project.RemoveMember)
			r.Get("/{id}/chat", a.handlers.Chat.GetChatByProjectId)
			r.Get("/{id}/chat/messages", a.handlers.Chat.ListMessagesByProjectId)
		})

		r.Route("/chats", func(r chi.Router) {
			r.Use(a.handlers.AuthMiddleware.ProtectRoutes)
			r.Post("/", a.handlers.Chat.GetOrCreateGeneralChat)
			r.Get("/", a.handlers.Chat.ListGeneralChats)
			r.Get("/{chatId}", a.handlers.Chat.GetChatById)
			r.Get("/{chatId}/messages", a.handlers.Chat.ListChatMessages)
			r.Get("/{chatId}/messages/{messageId}/reads", a.handlers.Chat.ListMessageReads)
			r.Post("/{chatId}/read", a.handlers.Chat.MarkChatRead)
			r.Post("/messages", a.handlers.Chat.CreateMessage)
		})

		r.Route("/tasks", func(r chi.Router) {
			r.Use(a.handlers.AuthMiddleware.ProtectRoutes)
			r.Get("/group-by-column", a.handlers.Task.GroupByColumn)
			r.Get("/count-by-column", a.handlers.Task.CountByColumn)
			r.Get("/", a.handlers.Task.List)
			r.Post("/", a.handlers.Task.Create)
			r.Get("/user", a.handlers.Task.ListUserDueTasks)
			r.Get("/{id}", a.handlers.Task.Get)
			r.Get("/{id}/comments", a.handlers.TaskComment.ListByTaskID)
			r.Post("/{id}/comments", a.handlers.TaskComment.Create)
			r.Post("/{id}/restore", a.handlers.Task.Restore)
			r.Put("/{id}", a.handlers.Task.Update)
			r.Delete("/{id}", a.handlers.Task.Archive)
			r.Patch("/{id}/move", a.handlers.Task.Move)
			r.Get("/search", a.handlers.Task.SearchTasksForUser)
		})

		r.Route("/notifications", func(r chi.Router) {
			r.Use(a.handlers.AuthMiddleware.ProtectRoutes)
			r.Get("/", a.handlers.Notification.List)
			r.Get("/unread-count", a.handlers.Notification.CountUnread)
			r.Post("/read-all", a.handlers.Notification.MarkAllRead)
			r.Post("/{id}/read", a.handlers.Notification.MarkRead)
		})

		r.Route("/ws", func(r chi.Router) {
			r.Get("/", a.Ws.Handler)
		})
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
