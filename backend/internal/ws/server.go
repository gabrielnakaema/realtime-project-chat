package ws

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type WsUser struct {
	id             uuid.UUID
	tokenExpiresAt time.Time
	writer         chan any
	reader         chan any
	rooms          map[uuid.UUID]bool
	lastPong       time.Time
	awaitingPong   bool
}

type WsRoomType string

const (
	WsRoomTypeChat    WsRoomType = "chat"
	WsRoomTypeProject WsRoomType = "project"
)

const (
	pingInterval        = 30 * time.Second
	pongTimeout         = 10 * time.Second
	usersOnlineInterval = 10 * time.Second
)

type WsRoom struct {
	id       uuid.UUID
	users    map[uuid.UUID]bool
	mutex    sync.Mutex
	roomType WsRoomType
}

type tokenProvider interface {
	Verify(token string) (*jwt.Token, error)
}

type chatService interface {
	GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Chat, error)
}

type projectService interface {
	GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Project, error)
}

type publisher interface {
	Publish(ctx context.Context, event events.Topic, data any) error
}

type Server struct {
	rooms          map[uuid.UUID]*WsRoom
	users          map[uuid.UUID]*WsUser
	logger         *slog.Logger
	mutex          sync.Mutex
	tokenProvider  tokenProvider
	chatService    chatService
	projectService projectService
	publisher      publisher
}

func NewServer(tokenProvider tokenProvider, logger *slog.Logger, chatService chatService, projectService projectService, publisher publisher) *Server {
	ws := &Server{
		rooms:          make(map[uuid.UUID]*WsRoom),
		logger:         logger,
		mutex:          sync.Mutex{},
		tokenProvider:  tokenProvider,
		chatService:    chatService,
		projectService: projectService,
		publisher:      publisher,
		users:          make(map[uuid.UUID]*WsUser),
	}

	go func() {
		usersOnlineTicker := time.NewTicker(usersOnlineInterval)

		for range usersOnlineTicker.C {
			for _, room := range ws.rooms {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

				userIds := []uuid.UUID{}
				for userId := range room.users {
					userIds = append(userIds, userId)
				}

				message := WebsocketMessage{
					Type:   WebsocketMessageTypeUsersOnline,
					RoomId: room.id,
					Data:   userIds,
				}

				ws.sendMessageToRoom(ctx, room.id, message)

				cancel()
			}
		}
	}()

	return ws
}

func (ws *Server) SendEvent(ctx context.Context, wsMessage WebsocketMessage) error {
	websocketMessage := WebsocketMessage{
		Type:   wsMessage.Type,
		RoomId: wsMessage.RoomId,
		Data:   wsMessage.Data,
	}

	return ws.sendMessageToRoom(ctx, wsMessage.RoomId, websocketMessage)
}

func (ws *Server) SendMessages(ctx context.Context, message *domain.ChatMessage) error {
	return ws.SendEvent(ctx, MapChatMessage(message))
}

func (ws *Server) SendUpdatedTask(ctx context.Context, task *domain.Task) error {
	return ws.SendEvent(ctx, MapTaskUpdated(task))
}

func (ws *Server) SendCreatedTask(ctx context.Context, task *domain.Task) error {
	return ws.SendEvent(ctx, MapTaskCreated(task))
}