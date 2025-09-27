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
	writer         chan interface{}
	reader         chan interface{}
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
	pingInterval = 30 * time.Second
	pongTimeout  = 10 * time.Second
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

type Server struct {
	rooms          map[uuid.UUID]*WsRoom
	users          map[uuid.UUID]*WsUser
	logger         *slog.Logger
	mutex          sync.Mutex
	tokenProvider  tokenProvider
	chatService    chatService
	projectService projectService
	eventMapper    *DomainEventMapper
}

type publisher interface {
	Publish(ctx context.Context, event events.Topic, data interface{}) error
}

func NewServer(tokenProvider tokenProvider, logger *slog.Logger, chatService chatService, projectService projectService, publisher publisher) *Server {
	return &Server{
		rooms:          make(map[uuid.UUID]*WsRoom),
		logger:         logger,
		mutex:          sync.Mutex{},
		tokenProvider:  tokenProvider,
		chatService:    chatService,
		projectService: projectService,
		users:          make(map[uuid.UUID]*WsUser),
		eventMapper:    &DomainEventMapper{},
	}
}

func (ws *Server) SendEvent(ctx context.Context, eventData EventData) error {
	websocketMessage := WebsocketMessage{
		Type:   eventData.Type,
		RoomId: eventData.RoomId,
		Data:   eventData.Data,
	}

	return ws.sendMessageToRoom(ctx, eventData.RoomId, websocketMessage)
}

func (ws *Server) SendMessages(ctx context.Context, message *domain.ChatMessage) error {
	return ws.SendEvent(ctx, ws.eventMapper.MapChatMessage(message))
}

func (ws *Server) SendUpdatedTask(ctx context.Context, task *domain.Task) error {
	return ws.SendEvent(ctx, ws.eventMapper.MapTaskUpdated(task))
}

func (ws *Server) SendCreatedTask(ctx context.Context, task *domain.Task) error {
	return ws.SendEvent(ctx, ws.eventMapper.MapTaskCreated(task))
}

func (ws *Server) sendMessageToRoom(ctx context.Context, roomId uuid.UUID, message WebsocketMessage) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	room, ok := ws.rooms[roomId]
	if !ok {
		return nil
	}

	for userId, _ := range room.users {
		user, ok := ws.users[userId]
		if !ok {
			continue
		}

		select {
		case user.writer <- message:
		case <-ctx.Done():
			return nil
		default:
			ws.logger.Error("failed to send message", "error", "channel is full", "user_id", user.id, "room_id", roomId)
			return nil
		}
	}

	return nil
}

func (ws *Server) disconnectUserFromRoom(userId uuid.UUID, roomId uuid.UUID) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	user, ok := ws.users[userId]
	if !ok {
		return
	}

	delete(user.rooms, roomId)

	room, ok := ws.rooms[roomId]
	if !ok {
		return
	}

	delete(room.users, userId)
	if len(room.users) == 0 {
		delete(ws.rooms, roomId)
	}
}

func (ws *Server) disconnectUser(userId uuid.UUID) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	user, ok := ws.users[userId]
	if !ok {
		return
	}

	for roomId, _ := range user.rooms {
		room, ok := ws.rooms[roomId]
		if !ok {
			continue
		}

		delete(room.users, userId)
		if len(room.users) == 0 {
			delete(ws.rooms, room.id)
		}
	}

	close(user.reader)
	close(user.writer)
	delete(ws.users, userId)
}

func (ws *Server) connectUserToRoom(userId uuid.UUID, roomId uuid.UUID, roomType WsRoomType) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	room, ok := ws.rooms[roomId]
	if !ok {
		room = &WsRoom{
			id:       roomId,
			users:    make(map[uuid.UUID]bool),
			mutex:    sync.Mutex{},
			roomType: roomType,
		}
		ws.rooms[roomId] = room
		room.users[userId] = true
		ws.users[userId].rooms[roomId] = true
	} else {
		room.users[userId] = true
		ws.users[userId].rooms[roomId] = true
	}
}
