package app

import (
	"github.com/gabrielnakaema/project-chat/internal/notification"
	"github.com/gabrielnakaema/project-chat/internal/platform/apphost"
	"github.com/gabrielnakaema/project-chat/internal/platform/auth"
	"github.com/gabrielnakaema/project-chat/internal/platform/postgres"
	"github.com/gabrielnakaema/project-chat/internal/platform/token"
)

type App struct {
	rt       *apphost.Runtime
	handlers *Handlers
}

type Handlers struct {
	AuthMiddleware *auth.Middleware
	Notification   *notification.Handler
}

func New() (*App, error) {
	rt, err := apphost.New("notification-service", "NOTIFICATION_SERVICE_PORT", "3335")
	if err != nil {
		return nil, err
	}
	pool, err := postgres.NewPool(rt.Config)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.TrackFunc(func() error {
		pool.Close()
		return nil
	})

	jwtProvider := token.NewTokenProvider(rt.Config)
	authMiddleware := auth.NewMiddleware(jwtProvider)

	notificationRepo := notification.NewRepository(pool)
	notificationService := notification.NewService(notificationRepo)
	notificationHandler := notification.NewHandler(notificationService)

	notificationSub, err := notification.NewEventSubscriber(rt.Ctx, rt.Config, rt.Logger, notificationRepo)
	if err != nil {
		rt.Close()
		return nil, err
	}
	rt.Track(notificationSub)

	app := App{
		rt: rt,
		handlers: &Handlers{
			AuthMiddleware: authMiddleware,
			Notification:   notificationHandler,
		},
	}

	return &app, nil
}

func (a *App) Close() {
	a.rt.Close()
}

func (a *App) Serve() error {
	return a.rt.Serve(a.Router())
}
