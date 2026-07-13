package api

import (
	"net/http"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/httpmw"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *Api) Router() http.Handler {
	r := chi.NewRouter()

	httpmw.Mount(r, a.rt.Config.CORSOrigins, a.rt.Logger)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Handle("/mcp", a.handlers.MCP)
	r.Handle("/health", httpmw.Health())

	r.Group(func(r chi.Router) {
		r.Use(a.handlers.AuthMiddleware.IdentifyUser)

		r.Route("/users", func(r chi.Router) {
			r.Post("/", a.handlers.User.Create)
			r.Get("/me", a.handlers.User.GetMe)
			r.Put("/me/password", a.handlers.User.ChangePassword)
			r.Get("/", a.handlers.User.ListUsers)
			r.Route("/me/mcp-api-keys", func(r chi.Router) {
				r.Use(a.handlers.AuthMiddleware.ProtectRoutes)
				r.Get("/scopes", a.handlers.MCPAPIKey.ListAvailableScopes)
				r.Post("/", a.handlers.MCPAPIKey.Create)
				r.Get("/", a.handlers.MCPAPIKey.List)
				r.Put("/{id}", a.handlers.MCPAPIKey.Update)
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
			r.Patch("/{id}/columns/{column_id}", a.handlers.Project.UpdateColumn)
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
			r.Get("/code-suggestions", a.handlers.Task.SuggestTaskCodes)
			r.Get("/project-search", a.handlers.Task.SearchProjectTasksForDependencies)
			r.Get("/{id}", a.handlers.Task.Get)
			r.Get("/{id}/comments", a.handlers.TaskComment.ListByTaskID)
			r.Post("/{id}/comments", a.handlers.TaskComment.Create)
			r.Post("/{id}/restore", a.handlers.Task.Restore)
			r.Put("/{id}", a.handlers.Task.Update)
			r.Delete("/{id}", a.handlers.Task.Archive)
			r.Patch("/{id}/move", a.handlers.Task.Move)
			r.Get("/search", a.handlers.Task.SearchTasksForUser)
		})
	})

	return r
}
