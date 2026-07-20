package realtime

import (
	"context"
	"errors"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/gabrielnakaema/project-chat/internal/platform/outbox"
	"github.com/google/uuid"
)

func (ws *Server) sendMessageToRoom(ctx context.Context, roomId uuid.UUID, message WebsocketMessage) error {
	ws.mutex.Lock()
	room, ok := ws.rooms[roomId]
	if !ok {
		ws.mutex.Unlock()
		return nil
	}

	userIds := make([]uuid.UUID, 0, len(room.users))
	for userId := range room.users {
		userIds = append(userIds, userId)
	}
	ws.mutex.Unlock()

	for _, userId := range userIds {
		ws.mutex.Lock()
		user, ok := ws.users[userId]
		ws.mutex.Unlock()
		if !ok {
			continue
		}

		select {
		case user.writer <- message:
		case <-ctx.Done():
			return nil
		default:
			ws.logger.Debug("failed to send message", "error", "channel is full", "user_id", user.id, "room_id", roomId)
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

	for roomId := range user.rooms {
		room, ok := ws.rooms[roomId]
		if !ok {
			continue
		}

		delete(room.users, userId)
		if len(room.users) == 0 {
			delete(ws.rooms, room.id)
		}
	}

	delete(ws.users, userId)
}

var errWebsocketUserDisconnected = errors.New("websocket user disconnected")

func (ws *Server) connectUserToRoom(ctx context.Context, userId uuid.UUID, roomId uuid.UUID, roomType WsRoomType) error {
	if !isValidRoomType(roomType) {
		return apperr.ValidationFailedError(map[string][]string{"room_type": {"invalid room type"}})
	}

	if roomType == WsRoomTypeChat || roomType == WsRoomTypeProject {
		authorizationCtx, cancel := context.WithTimeout(ctx, ws.authorizationTimeout)
		defer cancel()

		var err error
		switch roomType {
		case WsRoomTypeChat:
			err = ws.chatAuthorizer.CheckAccess(authorizationCtx, userId, roomId)
		case WsRoomTypeProject:
			err = ws.projectAuthorizer.CheckAccess(authorizationCtx, userId, roomId)
		}
		if err != nil {
			return err
		}
	}

	if roomType == WsRoomTypeUser {
		if roomId != userId {
			return apperr.ForbiddenError("forbidden")
		}
	}

	ws.mutex.Lock()

	user, ok := ws.users[userId]
	if !ok {
		ws.mutex.Unlock()
		return errWebsocketUserDisconnected
	}

	room, ok := ws.rooms[roomId]
	if ok {
		room.users[userId] = true
		user.rooms[roomId] = true
		ws.mutex.Unlock()
		ws.publishChatMemberViewed(ctx, userId, roomId, roomType)
		return nil
	}

	room = &WsRoom{
		id:       roomId,
		users:    make(map[uuid.UUID]bool),
		roomType: roomType,
	}
	ws.rooms[roomId] = room
	room.users[userId] = true
	user.rooms[roomId] = true
	ws.mutex.Unlock()
	ws.publishChatMemberViewed(ctx, userId, roomId, roomType)

	return nil
}

func (ws *Server) publishChatMemberViewed(ctx context.Context, userID uuid.UUID, roomID uuid.UUID, roomType WsRoomType) {
	if roomType != WsRoomTypeChat {
		return
	}

	chatMember := domain.ChatMember{
		ChatId:     roomID,
		UserId:     userID,
		LastSeenAt: time.Now(),
	}
	go func() {
		publishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := ws.enqueuer.Enqueue(publishCtx, outbox.Message{
			Topic:       events.ChatMemberViewed,
			AggregateID: roomID,
			Payload: &events.ChatMemberViewedPayload{
				ChatMember: chatMember,
				User:       domain.User{Id: userID},
			},
		}); err != nil {
			ws.logger.Error("failed to enqueue chat member viewed", "error", err.Error(), "user_id", userID, "room_id", roomID)
		}
	}()
}
