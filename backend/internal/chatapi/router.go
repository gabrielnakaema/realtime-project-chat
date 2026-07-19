package chatapi

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

		r.Route("/projects", func(r chi.Router) {
			r.Use(a.handlers.AuthMiddleware.ProtectRoutes)
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
	})

	return r
}
