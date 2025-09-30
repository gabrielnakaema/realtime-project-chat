package ws

import (
	"context"
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
	go ws.writerLoop(ctx, c, userId, writerChannel, &wg, cancel)

	wg.Add(1)
	go ws.readerLoop(ctx, c, userId, writerChannel, &wg, cancel)

	wg.Wait()
}

func (ws *Server) writerLoop(ctx context.Context, c *websocket.Conn, userId uuid.UUID, writerChannel chan interface{}, wg *sync.WaitGroup, cancel context.CancelFunc) {
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
}

func (ws *Server) readerLoop(ctx context.Context, c *websocket.Conn, userId uuid.UUID, writerChannel chan interface{}, wg *sync.WaitGroup, cancel context.CancelFunc) {
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
				}
				continue
			}

			ws.handleMessage(ctx, userId, message, writerChannel)
		}
	}
}