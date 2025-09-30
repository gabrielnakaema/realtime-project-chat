package ws

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

func (ws *Server) handleMessage(ctx context.Context, userId uuid.UUID, message WebsocketMessage, writerChannel chan interface{}) {
	switch message.Type {
	case WebsocketMessageTypeConnectUserToRoom:
		ws.handleConnectUserToRoom(ctx, userId, message)
	case WebsocketMessageTypeDisconnectUserFromRoom:
		ws.handleDisconnectUserFromRoom(ctx, userId, message)
	}
}

func (ws *Server) handleConnectUserToRoom(ctx context.Context, userId uuid.UUID, message WebsocketMessage) {
	bytes, err := json.Marshal(message.Data)
	if err != nil {
		ws.logger.Error("failed to marshal message", "error", err.Error(), "user_id", userId)
		return
	}

	var data ConnectUserToRoomData
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		ws.logger.Error("failed to unmarshal message", "error", err.Error(), "user_id", userId)
		return
	}

	err = ws.connectUserToRoom(userId, data.RoomId, WsRoomType(data.Type))
	if err != nil {
		return
	}

	ws.sendMessageToRoom(ctx, data.RoomId, WebsocketMessage{
		Type:   WebsocketMessageTypeUserConnected,
		RoomId: data.RoomId,
		Data: UserConnectedData{
			UserId: userId,
			RoomId: data.RoomId,
		},
	})
}

func (ws *Server) handleDisconnectUserFromRoom(ctx context.Context, userId uuid.UUID, message WebsocketMessage) {
	bytes, err := json.Marshal(message.Data)
	if err != nil {
		ws.logger.Error("failed to marshal message", "error", err.Error(), "user_id", userId)
		return
	}

	var data DisconnectUserFromRoomData
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		ws.logger.Error("failed to unmarshal message", "error", err.Error(), "user_id", userId)
		return
	}

	ws.disconnectUserFromRoom(userId, data.RoomId)
	ws.sendMessageToRoom(ctx, data.RoomId, WebsocketMessage{
		Type:   WebsocketMessageTypeUserDisconnected,
		RoomId: data.RoomId,
		Data: UserDisconnectedData{
			UserId: userId,
			RoomId: data.RoomId,
		},
	})
}