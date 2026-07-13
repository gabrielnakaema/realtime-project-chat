package websocketapi

import (
	"net/http"

	"github.com/gabrielnakaema/project-chat/internal/httpmw"
	"github.com/go-chi/chi/v5"
)

func (a *App) Router() http.Handler {
	r := chi.NewRouter()

	httpmw.Mount(r, a.rt.Config.CORSOrigins, a.rt.Logger)

	r.Get("/ws", a.Ws.Handler)
	r.Get("/health", httpmw.Health())

	return r
}
