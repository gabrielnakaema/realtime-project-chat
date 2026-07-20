package app

import (
	"net/http"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/platform/httpmw"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *App) Router() http.Handler {
	r := chi.NewRouter()

	httpmw.Mount(r, a.rt.Config.CORSOrigins, a.rt.Logger)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Handle("/health", httpmw.Health())
	r.Handle("/mcp", a.handler)

	return r
}
