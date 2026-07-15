package ws

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/outbox"
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
	WsRoomTypeUser    WsRoomType = "user"
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

type accessChecker interface {
	CheckAccess(ctx context.Context, userID uuid.UUID, resourceID uuid.UUID) error
}

type outboxEnqueuer interface {
	Enqueue(ctx context.Context, msgs ...outbox.Message) error
}

type Server struct {
	rooms                map[uuid.UUID]*WsRoom
	users                map[uuid.UUID]*WsUser
	logger               *slog.Logger
	mutex                sync.Mutex
	tokenProvider        tokenProvider
	chatAuthorizer       accessChecker
	projectAuthorizer    accessChecker
	authorizationTimeout time.Duration
	enqueuer             outboxEnqueuer
}

func NewServer(ctx context.Context, tokenProvider tokenProvider, logger *slog.Logger, chatAuthorizer accessChecker, projectAuthorizer accessChecker, authorizationTimeout time.Duration, enqueuer outboxEnqueuer) *Server {
	if authorizationTimeout <= 0 {
		authorizationTimeout = 2 * time.Second
	}
	ws := &Server{
		rooms:                make(map[uuid.UUID]*WsRoom),
		logger:               logger,
		mutex:                sync.Mutex{},
		tokenProvider:        tokenProvider,
		chatAuthorizer:       chatAuthorizer,
		projectAuthorizer:    projectAuthorizer,
		authorizationTimeout: authorizationTimeout,
		enqueuer:             enqueuer,
		users:                make(map[uuid.UUID]*WsUser),
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

func (ws *Server) SendUpdatedTask(ctx context.Context, task *domain.Task, previousProjectColumnID *uuid.UUID) error {
	return ws.SendEvent(ctx, MapTaskUpdated(task, previousProjectColumnID))
}

func (ws *Server) SendUpdatedProject(ctx context.Context, project *domain.Project) error {
	return ws.SendEvent(ctx, MapProjectUpdated(project))
}

func (ws *Server) SendProjectMemberCreated(ctx context.Context, member *domain.ProjectMember) error {
	if err := ws.SendEvent(ctx, MapProjectMemberCreated(member, member.UserId)); err != nil {
		return err
	}

	return ws.SendEvent(ctx, MapProjectMemberCreated(member, member.ProjectId))
}

func (ws *Server) SendProjectMemberRemoved(ctx context.Context, member *domain.ProjectMember) error {
	if err := ws.SendEvent(ctx, MapProjectMemberRemoved(member, member.UserId)); err != nil {
		return err
	}
	if err := ws.SendEvent(ctx, MapProjectMemberRemoved(member, member.ProjectId)); err != nil {
		return err
	}

	ws.disconnectUserFromRoom(member.UserId, member.ProjectId)
	return nil
}

func (ws *Server) SendChatProjectMemberCreated(ctx context.Context, member *domain.ProjectMember, chatID uuid.UUID) error {
	return ws.SendEvent(ctx, MapProjectMemberCreated(member, chatID))
}

func (ws *Server) SendChatProjectMemberRemoved(ctx context.Context, member *domain.ProjectMember, chatID uuid.UUID) error {
	if err := ws.SendEvent(ctx, MapProjectMemberRemoved(member, chatID)); err != nil {
		return err
	}

	ws.disconnectUserFromRoom(member.UserId, chatID)
	return nil
}

func (ws *Server) SendCreatedTask(ctx context.Context, task *domain.Task) error {
	return ws.SendEvent(ctx, MapTaskCreated(task))
}

func (ws *Server) SendCreatedTaskComment(ctx context.Context, comment *domain.TaskComment) error {
	return ws.SendEvent(ctx, MapTaskCommentCreated(comment))
}

func (ws *Server) SendNotification(ctx context.Context, notification *domain.Notification) error {
	return ws.SendEvent(ctx, MapNotificationCreated(notification))
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
		if room.roomType == WsRoomTypeUser {
			continue
		}

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
	case WsRoomTypeChat, WsRoomTypeProject, WsRoomTypeUser:
		return true
	default:
		return false
	}
}
