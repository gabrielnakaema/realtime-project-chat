package ws

import (
	"context"
	"sync"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/google/uuid"
)

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