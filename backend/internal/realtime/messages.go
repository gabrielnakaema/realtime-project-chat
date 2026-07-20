package realtime

import (
	"context"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gabrielnakaema/project-chat/internal/domain"
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
	WebsocketMessageTypeTaskCreated            WebsocketMessageType = "task_created"
	WebsocketMessageTypeTaskUpdated            WebsocketMessageType = "task_updated"
	WebsocketMessageTypeProjectUpdated         WebsocketMessageType = "project_updated"
	WebsocketMessageTypeProjectMemberCreated   WebsocketMessageType = "project_member_created"
	WebsocketMessageTypeProjectMemberRemoved   WebsocketMessageType = "project_member_removed"
	WebsocketMessageTypeTaskCommentCreated     WebsocketMessageType = "task_comment_created"
	WebsocketMessageTypeUsersOnline            WebsocketMessageType = "users_online"
	WebsocketMessageTypeMessageRead            WebsocketMessageType = "message_read"
	WebsocketMessageTypeNotificationCreated    WebsocketMessageType = "notification_created"
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

func MapChatMessage(message *domain.ChatMessage) WebsocketMessage {
	return WebsocketMessage{
		Type:   WebsocketMessageTypeMessage,
		RoomId: message.ChatId,
		Data:   message,
	}
}

func MapChatMessageRead(read *domain.ChatMessageRead, chatId uuid.UUID) WebsocketMessage {
	return WebsocketMessage{
		Type:   WebsocketMessageTypeMessageRead,
		RoomId: chatId,
		Data:   read,
	}
}

func MapTaskCreated(task *domain.Task) WebsocketMessage {
	return WebsocketMessage{
		Type:   WebsocketMessageTypeTaskCreated,
		RoomId: task.ProjectId,
		Data:   task,
	}
}

type TaskUpdatedData struct {
	Task                    *domain.Task `json:"task"`
	PreviousProjectColumnID *uuid.UUID   `json:"previous_project_column_id,omitempty"`
}

func MapTaskUpdated(task *domain.Task, previousProjectColumnID *uuid.UUID) WebsocketMessage {
	return WebsocketMessage{
		Type:   WebsocketMessageTypeTaskUpdated,
		RoomId: task.ProjectId,
		Data: TaskUpdatedData{
			Task:                    task,
			PreviousProjectColumnID: previousProjectColumnID,
		},
	}
}

func MapProjectUpdated(project *domain.Project) WebsocketMessage {
	return WebsocketMessage{
		Type:   WebsocketMessageTypeProjectUpdated,
		RoomId: project.Id,
		Data:   project,
	}
}

func MapProjectMemberCreated(member *domain.ProjectMember, roomID uuid.UUID) WebsocketMessage {
	return WebsocketMessage{
		Type:   WebsocketMessageTypeProjectMemberCreated,
		RoomId: roomID,
		Data:   member,
	}
}

func MapProjectMemberRemoved(member *domain.ProjectMember, roomID uuid.UUID) WebsocketMessage {
	return WebsocketMessage{
		Type:   WebsocketMessageTypeProjectMemberRemoved,
		RoomId: roomID,
		Data:   member,
	}
}

func MapTaskCommentCreated(comment *domain.TaskComment) WebsocketMessage {
	return WebsocketMessage{
		Type:   WebsocketMessageTypeTaskCommentCreated,
		RoomId: comment.Task.ProjectId,
		Data:   comment,
	}
}

func MapNotificationCreated(notification *domain.Notification) WebsocketMessage {
	return WebsocketMessage{
		Type:   WebsocketMessageTypeNotificationCreated,
		RoomId: notification.UserId,
		Data:   notification,
	}
}
