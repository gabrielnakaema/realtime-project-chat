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
	Publish(ctx context.Context, event events.Topic, data events.Payload) error
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

func NewServer(ctx context.Context, tokenProvider tokenProvider, logger *slog.Logger, chatService chatService, projectService projectService, publisher publisher) *Server {
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
		defer usersOnlineTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Info("websocket server ticker stopped")
				return
			case <-usersOnlineTicker.C:
				for _, room := range ws.snapshotRooms() {
					tickCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
					message := WebsocketMessage{
						Type:   WebsocketMessageTypeUsersOnline,
						RoomId: room.id,
						Data:   room.userIds,
					}

					ws.sendMessageToRoom(tickCtx, room.id, message)
					cancel()
				}
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

func (ws *Server) SendReadUpdate(ctx context.Context, chatId uuid.UUID, read *domain.ChatMessageRead) error {
	return ws.SendEvent(ctx, MapChatMessageRead(read, chatId))
}

func (ws *Server) SendUpdatedTask(ctx context.Context, task *domain.Task, previousStatus *domain.TaskStatus) error {
	return ws.SendEvent(ctx, MapTaskUpdated(task, previousStatus))
}

func (ws *Server) SendCreatedTask(ctx context.Context, task *domain.Task) error {
	return ws.SendEvent(ctx, MapTaskCreated(task))
}

func (ws *Server) SendCreatedTaskComment(ctx context.Context, comment *domain.TaskComment) error {
	return ws.SendEvent(ctx, MapTaskCommentCreated(comment))
}

type roomSnapshot struct {
	id      uuid.UUID
	userIds []uuid.UUID
}

func (ws *Server) snapshotRooms() []roomSnapshot {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	rooms := make([]roomSnapshot, 0, len(ws.rooms))
	for _, room := range ws.rooms {
		userIds := make([]uuid.UUID, 0, len(room.users))
		for userId := range room.users {
			userIds = append(userIds, userId)
		}

		rooms = append(rooms, roomSnapshot{
			id:      room.id,
			userIds: userIds,
		})
	}

	return rooms
}

func isValidRoomType(roomType WsRoomType) bool {
	switch roomType {
	case WsRoomTypeChat, WsRoomTypeProject:
		return true
	default:
		return false
	}
}
