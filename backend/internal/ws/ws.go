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

func (ws *Server) sendMessageToRoom(ctx context.Context, roomId uuid.UUID, message WebsocketMessage) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	room, ok := ws.rooms[roomId]
	if !ok {
		return nil
	}

	for userId := range room.users {
		user, ok := ws.users[userId]
		if !ok {
			continue
		}

		select {
		case user.writer <- message:
		case <-ctx.Done():
			return nil
		default:
			ws.logger.Debug("failed to send message", "error", "channel is full", "user_id", user.id, "room_id", roomId)
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

func (ws *Server) connectUserToRoom(userId uuid.UUID, roomId uuid.UUID, roomType WsRoomType) error {
	if roomType == WsRoomTypeChat {
		_, err := ws.chatService.GetById(context.Background(), roomId, userId)
		if err != nil {
			return err
		}

		chatMember := domain.ChatMember{
			ChatId:     roomId,
			UserId:     userId,
			LastSeenAt: time.Now(),
		}

		go func() {
			ws.publisher.Publish(context.Background(), events.ChatMemberViewed, chatMember)
		}()
	}

	if roomType == WsRoomTypeProject {
		_, err := ws.projectService.GetById(context.Background(), roomId, userId)
		if err != nil {
			return err
		}
	}

	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	room, ok := ws.rooms[roomId]
	if ok {
		room.users[userId] = true
		ws.users[userId].rooms[roomId] = true
		return nil
	}

	room = &WsRoom{
		id:       roomId,
		users:    make(map[uuid.UUID]bool),
		mutex:    sync.Mutex{},
		roomType: roomType,
	}
	ws.rooms[roomId] = room
	room.users[userId] = true
	ws.users[userId].rooms[roomId] = true

	return nil
}
