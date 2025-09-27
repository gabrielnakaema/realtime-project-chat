package ws

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

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

	token, err := ws.tokenProvider.Verify(authorization)
	if err != nil {
		WriteErrorAndClose(r.Context(), c, "invalid token")
		return
	}

	sub := token.Claims.(jwt.MapClaims)["sub"].(string)
	userId, err := uuid.Parse(sub)
	if err != nil {
		WriteErrorAndClose(r.Context(), c, "invalid user_id")
		return
	}

	writerChannel := make(chan interface{})
	readerChannel := make(chan interface{})
	expiresAt := token.Claims.(jwt.MapClaims)["exp"].(float64)

	roomUser := &WsUser{
		id:             userId,
		tokenExpiresAt: time.Unix(int64(expiresAt), 0),
		writer:         writerChannel,
		reader:         readerChannel,
		rooms:          make(map[uuid.UUID]bool),
		lastPong:       time.Now(),
		awaitingPong:   false,
	}

	ws.mutex.Lock()
	ws.users[userId] = roomUser
	ws.mutex.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	cleanUp := func() {
		c.Close(websocket.StatusNormalClosure, "close")
		ws.disconnectUser(userId)
		cancel()
	}

	defer cleanUp()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		pingTicker := time.NewTicker(pingInterval)

		defer func() {
			pingTicker.Stop()
			wg.Done()
			cancel()
		}()

		for {
			select {
			case message, ok := <-writerChannel:
				if !ok {
					return
				}

				err := WriteWebsocketMessage(ctx, c, message.(WebsocketMessage))
				if err != nil {
					ws.logger.Error("failed to send message", "error", err.Error(), "user_id", userId)
					return
				}
			case <-pingTicker.C:
				ws.mutex.Lock()
				user := ws.users[userId]
				if user == nil {
					ws.mutex.Unlock()
					return
				}

				if user.awaitingPong && time.Since(user.lastPong) > pongTimeout {
					ws.mutex.Unlock()
					return
				}

				user.awaitingPong = true
				ws.mutex.Unlock()

				pingMessage := WebsocketMessage{
					Type: WebsocketMessageTypePing,
					Data: nil,
				}
				err := WriteWebsocketMessage(ctx, c, pingMessage)
				if err != nil {
					ws.logger.Error("failed to send ping", "error", err.Error(), "user_id", userId)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer func() {
			cancel()
			wg.Done()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				c.SetReadLimit(32768)
				var message WebsocketMessage
				err := wsjson.Read(ctx, c, &message)
				if err != nil {
					var websocketErr *websocket.CloseError
					if errors.As(err, &websocketErr) {
						return
					}
					ws.logger.Error("failed to read message", "error", err.Error(), "user_id", userId)
					return
				}

				if message.Type == WebsocketMessageTypePong {
					ws.mutex.Lock()
					user := ws.users[userId]
					if user != nil {
						user.lastPong = time.Now()
						user.awaitingPong = false
					}
					ws.mutex.Unlock()
					continue
				}

				if message.Type == WebsocketMessageTypePing {
					pongMessage := WebsocketMessage{
						Type: WebsocketMessageTypePong,
						Data: nil,
					}
					select {
					case writerChannel <- pongMessage:
					default:
						ws.logger.Error("failed to send pong response", "error", "writer channel full", "user_id", userId)
					}
					continue
				}

				if message.Type == WebsocketMessageTypeConnectUserToRoom {
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
					continue
				}

				if message.Type == WebsocketMessageTypeDisconnectUserFromRoom {

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
					continue
				}
			}
		}
	}()

	wg.Wait()
}
