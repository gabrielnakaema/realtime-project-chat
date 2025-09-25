package ws

import (
	"context"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
)

type ErrorMessage struct {
	Message string `json:"message"`
}

type WebsocketMessageType string

const (
	WebsocketMessageTypeError                  WebsocketMessageType = "error"
	WebsocketMessageTypeMessage                WebsocketMessageType = "message"
	WebsocketMessageTypeUserDisconnected       WebsocketMessageType = "user_disconnected"
	WebsocketMessageTypeUserConnected          WebsocketMessageType = "user_connected"
	WebsocketMessageTypePing                   WebsocketMessageType = "ping"
	WebsocketMessageTypePong                   WebsocketMessageType = "pong"
	WebsocketMessageTypeConnectUserToRoom      WebsocketMessageType = "connect_user_to_room"
	WebsocketMessageTypeDisconnectUserFromRoom WebsocketMessageType = "disconnect_user_from_room"
)

type WebsocketMessage struct {
	Type   WebsocketMessageType `json:"type"`
	RoomId uuid.UUID            `json:"room_id"`
	Data   interface{}          `json:"data"`
}

func WriteErrorMessage(ctx context.Context, c *websocket.Conn, text string) error {
	message := ErrorMessage{
		Message: text,
	}

	return wsjson.Write(ctx, c, WebsocketMessage{
		Type: WebsocketMessageTypeError,
		Data: message,
	})
}

func WriteWebsocketMessage(ctx context.Context, c *websocket.Conn, data WebsocketMessage) error {
	return wsjson.Write(ctx, c, data)
}

func WriteErrorAndClose(ctx context.Context, c *websocket.Conn, text string) error {
	err := WriteErrorMessage(ctx, c, text)
	if err != nil {
		c.Close(websocket.StatusInternalError, text)
		return err
	}

	return c.Close(websocket.StatusNormalClosure, text)
}

type ConnectUserToRoomData struct {
	RoomId uuid.UUID  `json:"room_id"`
	Type   WsRoomType `json:"type"`
}

type DisconnectUserFromRoomData struct {
	RoomId uuid.UUID `json:"room_id"`
}

type UserConnectedData struct {
	UserId uuid.UUID `json:"user_id"`
	RoomId uuid.UUID `json:"room_id"`
}

type UserDisconnectedData struct {
	UserId uuid.UUID `json:"user_id"`
	RoomId uuid.UUID `json:"room_id"`
}
