package chat_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/auth"
	"github.com/gabrielnakaema/project-chat/internal/chat"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockChatService struct {
	mock.Mock
}

func (m *mockChatService) GetByProjectId(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Chat, error) {
	args := m.Called(ctx, id, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Chat), args.Error(1)
}

func (m *mockChatService) GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Chat, error) {
	args := m.Called(ctx, id, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Chat), args.Error(1)
}

func (m *mockChatService) CreateMessage(ctx context.Context, request chat.CreateChatMessageRequest) (*domain.ChatMessage, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ChatMessage), args.Error(1)
}

func (m *mockChatService) ListMessagesByProjectId(ctx context.Context, request chat.ListMessagesByProjectIdRequest) (*utils.CursorPaginated[domain.ChatMessage], error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.CursorPaginated[domain.ChatMessage]), args.Error(1)
}

func (m *mockChatService) GetOrCreateGeneralChat(ctx context.Context, currentUserId uuid.UUID, targetUserIds []uuid.UUID) (*domain.Chat, error) {
	args := m.Called(ctx, currentUserId, targetUserIds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Chat), args.Error(1)
}

func (m *mockChatService) ListGeneralChats(ctx context.Context, userId uuid.UUID) ([]domain.Chat, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Chat), args.Error(1)
}

func (m *mockChatService) ListMessagesByChatId(ctx context.Context, request chat.ListMessagesByChatIdRequest) (*utils.CursorPaginated[domain.ChatMessage], error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.CursorPaginated[domain.ChatMessage]), args.Error(1)
}

func (m *mockChatService) MarkChatRead(ctx context.Context, request chat.MarkChatReadRequest) error {
	return m.Called(ctx, request).Error(0)
}

func (m *mockChatService) ListMessageReads(ctx context.Context, request chat.ListMessageReadsRequest) ([]domain.ChatMessageRead, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ChatMessageRead), args.Error(1)
}

func withUserId(req *http.Request, userId uuid.UUID) *http.Request {
	ctx := context.WithValue(req.Context(), auth.UserIdContextKey, userId)
	return req.WithContext(ctx)
}

func withChiParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func encodeJSON(t *testing.T, v interface{}) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		t.Fatal(err)
	}
	return &buf
}

func TestChatHandler_GetOrCreateGeneralChat(t *testing.T) {
	userId := uuid.New()
	otherUserId := uuid.New()

	validChat := &domain.Chat{
		Id:       uuid.New(),
		ChatType: domain.ChatTypeGeneral,
		Members: []domain.ChatMember{
			{UserId: userId},
			{UserId: otherUserId},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		userId         uuid.UUID
		mockSetup      func(*mockChatService)
		expectedStatus int
	}{
		{
			name:           "invalid JSON body returns 400",
			requestBody:    `{"user_ids":}`,
			userId:         userId,
			mockSetup:      func(m *mockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty user_ids returns 422",
			requestBody:    chat.GetOrCreateGeneralChatRequest{UserIds: []uuid.UUID{}},
			userId:         userId,
			mockSetup:      func(m *mockChatService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "nil UUID in user_ids returns 422",
			requestBody:    chat.GetOrCreateGeneralChatRequest{UserIds: []uuid.UUID{uuid.Nil}},
			userId:         userId,
			mockSetup:      func(m *mockChatService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:        "valid request returns 200 with chat",
			requestBody: chat.GetOrCreateGeneralChatRequest{UserIds: []uuid.UUID{otherUserId}},
			userId:      userId,
			mockSetup: func(m *mockChatService) {
				m.On("GetOrCreateGeneralChat", mock.Anything, userId, []uuid.UUID{otherUserId}).
					Return(validChat, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "service not found returns 404",
			requestBody: chat.GetOrCreateGeneralChatRequest{UserIds: []uuid.UUID{otherUserId}},
			userId:      userId,
			mockSetup: func(m *mockChatService) {
				m.On("GetOrCreateGeneralChat", mock.Anything, userId, []uuid.UUID{otherUserId}).
					Return(nil, domain.NotFoundError("user not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockChatService{}
			tt.mockSetup(svc)
			handler := chat.NewHandler(svc)

			var body *bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body = bytes.NewBufferString(str)
			} else {
				body = encodeJSON(t, tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/chats/", body)
			req.Header.Set("Content-Type", "application/json")
			req = withUserId(req, tt.userId)
			w := httptest.NewRecorder()

			handler.GetOrCreateGeneralChat(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestChatHandler_GetChatById(t *testing.T) {
	userId := uuid.New()
	chatId := uuid.New()

	validChat := &domain.Chat{
		Id:        chatId,
		ChatType:  domain.ChatTypeGeneral,
		Members:   []domain.ChatMember{{UserId: userId}},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name           string
		chatId         string
		userId         uuid.UUID
		mockSetup      func(*mockChatService)
		expectedStatus int
	}{
		{
			name:           "missing chatId returns 400",
			chatId:         "",
			userId:         userId,
			mockSetup:      func(m *mockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid chatId returns 400",
			chatId:         "not-a-uuid",
			userId:         userId,
			mockSetup:      func(m *mockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing user returns 401",
			chatId:         chatId.String(),
			userId:         uuid.Nil,
			mockSetup:      func(m *mockChatService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "valid request returns 200",
			chatId: chatId.String(),
			userId: userId,
			mockSetup: func(m *mockChatService) {
				m.On("GetById", mock.Anything, chatId, userId).Return(validChat, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "service not found returns 404",
			chatId: chatId.String(),
			userId: userId,
			mockSetup: func(m *mockChatService) {
				m.On("GetById", mock.Anything, chatId, userId).Return(nil, domain.NotFoundError("chat not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockChatService{}
			tt.mockSetup(svc)
			handler := chat.NewHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/chats/"+tt.chatId, nil)
			req = withUserId(req, tt.userId)
			req = withChiParam(req, "chatId", tt.chatId)

			w := httptest.NewRecorder()
			handler.GetChatById(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestChatHandler_ListChatMessages(t *testing.T) {
	userId := uuid.New()
	chatId := uuid.New()

	emptyResult := &utils.CursorPaginated[domain.ChatMessage]{
		Data:    []domain.ChatMessage{},
		HasNext: false,
	}

	tests := []struct {
		name           string
		chatId         string
		queryParams    map[string]string
		userId         uuid.UUID
		mockSetup      func(*mockChatService)
		expectedStatus int
	}{
		{
			name:           "invalid chatId returns 400",
			chatId:         "not-a-uuid",
			queryParams:    map[string]string{},
			userId:         userId,
			mockSetup:      func(m *mockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "limit=0 returns 400",
			chatId:         chatId.String(),
			queryParams:    map[string]string{"limit": "0"},
			userId:         userId,
			mockSetup:      func(m *mockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "limit=51 returns 400",
			chatId:         chatId.String(),
			queryParams:    map[string]string{"limit": "51"},
			userId:         userId,
			mockSetup:      func(m *mockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid before time format returns 400",
			chatId:         chatId.String(),
			queryParams:    map[string]string{"before": "not-a-date"},
			userId:         userId,
			mockSetup:      func(m *mockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid id param returns 400",
			chatId:         chatId.String(),
			queryParams:    map[string]string{"id": "not-a-uuid"},
			userId:         userId,
			mockSetup:      func(m *mockChatService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "valid request returns 200",
			chatId:      chatId.String(),
			queryParams: map[string]string{},
			userId:      userId,
			mockSetup: func(m *mockChatService) {
				m.On("ListMessagesByChatId", mock.Anything, mock.MatchedBy(func(req chat.ListMessagesByChatIdRequest) bool {
					return req.ChatId == chatId && req.UserId == userId
				})).Return(emptyResult, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "valid request with before param returns 200",
			chatId:      chatId.String(),
			queryParams: map[string]string{"before": "2026-01-01T00:00:00Z", "limit": "5"},
			userId:      userId,
			mockSetup: func(m *mockChatService) {
				m.On("ListMessagesByChatId", mock.Anything, mock.MatchedBy(func(req chat.ListMessagesByChatIdRequest) bool {
					return req.ChatId == chatId && req.UserId == userId && req.Params.Limit == 5
				})).Return(emptyResult, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "service forbidden returns 403",
			chatId:      chatId.String(),
			queryParams: map[string]string{},
			userId:      userId,
			mockSetup: func(m *mockChatService) {
				m.On("ListMessagesByChatId", mock.Anything, mock.Anything).
					Return(nil, domain.ForbiddenError("forbidden"))
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockChatService{}
			tt.mockSetup(svc)
			handler := chat.NewHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/chats/"+tt.chatId+"/messages", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Set(k, v)
			}
			req.URL.RawQuery = q.Encode()

			req = withUserId(req, tt.userId)
			req = withChiParam(req, "chatId", tt.chatId)

			w := httptest.NewRecorder()
			handler.ListChatMessages(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			svc.AssertExpectations(t)
		})
	}
}
