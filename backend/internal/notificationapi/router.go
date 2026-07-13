package notificationapi

import (
	"net/http"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/httpmw"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *App) Router() http.Handler {
	r := chi.NewRouter()

	httpmw.Mount(r, a.rt.Config.CORSOrigins, a.rt.Logger)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Handle("/health", httpmw.Health())

	r.Group(func(r chi.Router) {
		r.Use(a.handlers.AuthMiddleware.IdentifyUser)

		r.Route("/notifications", func(r chi.Router) {
			r.Use(a.handlers.AuthMiddleware.ProtectRoutes)
			r.Get("/", a.handlers.Notification.List)
			r.Get("/unread-count", a.handlers.Notification.CountUnread)
			r.Post("/read-all", a.handlers.Notification.MarkAllRead)
			r.Post("/{id}/read", a.handlers.Notification.MarkRead)
		})
	})

	return r
}
