package ws

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ChatRoomUser struct {
	Id             uuid.UUID
	TokenExpiresAt time.Time
	channel        chan interface{}
}

type ChatRoom struct {
	Id    uuid.UUID
	Users map[uuid.UUID]*ChatRoomUser
}

type tokenProvider interface {
	Verify(token string) (*jwt.Token, error)
}

type chatRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*domain.Chat, error)
}

type Server struct {
	chatRooms      map[uuid.UUID]*ChatRoom
	receiver       chan interface{}
	logger         *slog.Logger
	mutex          sync.Mutex
	tokenProvider  tokenProvider
	chatRepository chatRepository
}

type publisher interface {
	Publish(ctx context.Context, event events.Topic, data interface{}) error
}

func NewServer(tokenProvider tokenProvider, logger *slog.Logger, chatRepository chatRepository, publisher publisher) *Server {
	return &Server{
		chatRooms:      make(map[uuid.UUID]*ChatRoom),
		receiver:       make(chan interface{}),
		logger:         logger,
		mutex:          sync.Mutex{},
		tokenProvider:  tokenProvider,
		chatRepository: chatRepository,
	}
}

type ErrorMessage struct {
	Message string `json:"message"`
}

type WebsocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func WriteErrorMessage(ctx context.Context, c *websocket.Conn, text string) error {
	message := ErrorMessage{
		Message: text,
	}

	return wsjson.Write(ctx, c, WebsocketMessage{
		Type: "error",
		Data: message,
	})
}

func WriteWebsocketMessage(ctx context.Context, c *websocket.Conn, data interface{}) error {
	return wsjson.Write(ctx, c, WebsocketMessage{
		Type: "message",
		Data: data,
	})
}

func WriteErrorAndClose(ctx context.Context, c *websocket.Conn, text string) error {
	err := WriteErrorMessage(ctx, c, text)
	if err != nil {
		c.Close(websocket.StatusInternalError, text)
		return err
	}

	return c.Close(websocket.StatusNormalClosure, text)
}

func (ws *Server) Handler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		ws.logger.Error("failed to accept websocket", "error", err.Error())
		return
	}
	defer c.Close(websocket.StatusNormalClosure, "close")

	authorization := r.URL.Query().Get("jwt")
	if authorization == "" {
		WriteErrorAndClose(r.Context(), c, "jwt is required")
		return
	}

	chatId := r.URL.Query().Get("chat_id")
	if chatId == "" {
		WriteErrorAndClose(r.Context(), c, "chat_id is required")
		return
	}

	token, err := ws.tokenProvider.Verify(authorization)
	if err != nil {
		WriteErrorAndClose(r.Context(), c, "invalid token")
		return
	}

	userId := token.Claims.(jwt.MapClaims)["sub"].(string)

	chatRoomId, err := uuid.Parse(chatId)
	if err != nil {
		WriteErrorAndClose(r.Context(), c, "invalid chat_id")
		return
	}

	chat, err := ws.chatRepository.GetById(r.Context(), chatRoomId)
	if err != nil {
		WriteErrorAndClose(r.Context(), c, "chat not found")
		return
	}

	hasPermission := false
	for _, member := range chat.Members {
		if member.UserId == uuid.MustParse(userId) {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		WriteErrorAndClose(r.Context(), c, "user not allowed to join chat")
		return
	}

	writeChannel := make(chan interface{})
	userUuid, err := uuid.Parse(userId)
	if err != nil {
		WriteErrorAndClose(r.Context(), c, "invalid user_id")
		return
	}

	expiresAt := token.Claims.(jwt.MapClaims)["exp"].(float64)

	chatRoomUser := &ChatRoomUser{
		Id:             userUuid,
		TokenExpiresAt: time.Unix(int64(expiresAt), 0),
		channel:        writeChannel,
	}

	ws.mutex.Lock()

	_, ok := ws.chatRooms[chatRoomId]
	if !ok {
		ws.chatRooms[chatRoomId] = &ChatRoom{
			Id:    chatRoomId,
			Users: make(map[uuid.UUID]*ChatRoomUser),
		}
	}

	ws.chatRooms[chatRoomId].Users[userUuid] = chatRoomUser

	ws.mutex.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})

	cleanUp := func() {
		cancel()

		ws.mutex.Lock()
		if room, exists := ws.chatRooms[chatRoomId]; exists {
			delete(room.Users, userUuid)
			if len(room.Users) == 0 {
				delete(ws.chatRooms, chatRoomId)
			}
		}
		ws.mutex.Unlock()

		close(writeChannel)
	}

	defer cleanUp()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case message, ok := <-writeChannel:
				if !ok {
					return
				}

				err := WriteWebsocketMessage(ctx, c, message)
				if err != nil {
					ws.logger.Error("failed to send message", "error", err.Error(), "user_id", userId, "chat_room_id", chatRoomId)
					return
				}
			case <-done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			close(done)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				c.SetReadLimit(32768)
				_, message, err := c.Read(ctx)
				if err != nil {
					var websocketErr *websocket.CloseError
					if errors.As(err, &websocketErr) {
						return
					}
					ws.logger.Error("failed to read message", "error", err.Error(), "user_id", userId, "chat_room_id", chatRoomId)
					return
				}

				ws.logger.Info("message", "receiving message from ws client", message)
			}
		}
	}()

	wg.Wait()
}

func (ws *Server) SendMessages(ctx context.Context, message *domain.ChatMessage) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	chatRoom, ok := ws.chatRooms[message.ChatId]
	if !ok {
		return nil
	}

	for _, user := range chatRoom.Users {
		select {
		case user.channel <- message:
		case <-ctx.Done():
			return nil
		default:
			ws.logger.Error("failed to send message", "error", "channel is full", "user_id", user.Id, "chat_room_id", message.ChatId)
			return nil
		}
	}

	return nil
}
