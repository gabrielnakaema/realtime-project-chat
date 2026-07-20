package tasksapi

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
