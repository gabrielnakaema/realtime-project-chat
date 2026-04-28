package ws

import (
	"context"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
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

func (ws *Server) connectUserToRoom(userId uuid.UUID, roomId uuid.UUID, roomType WsRoomType) error {
	if !isValidRoomType(roomType) {
		return nil
	}

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
			ws.publisher.Publish(context.Background(), events.ChatMemberViewed, &events.ChatMemberViewedPayload{
				ChatMember: chatMember,
				User: domain.User{
					Id: userId,
				},
			})
		}()
	}

	if roomType == WsRoomTypeProject {
		_, err := ws.projectService.GetById(context.Background(), roomId, userId)
		if err != nil {
			return err
		}
	}

	if roomType == WsRoomTypeUser {
		if roomId != userId {
			return domain.ForbiddenError("forbidden")
		}
	}

	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	user, ok := ws.users[userId]
	if !ok {
		return nil
	}

	room, ok := ws.rooms[roomId]
	if ok {
		room.users[userId] = true
		user.rooms[roomId] = true
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

	return nil
}
