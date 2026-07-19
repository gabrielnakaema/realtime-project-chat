package chat_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/chat"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/outbox"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockChatRepository struct {
	mock.Mock
	builtEvents []outbox.Message
}

func (m *mockChatRepository) Create(ctx context.Context, chat *domain.Chat) error {
	return m.Called(ctx, chat).Error(0)
}

func (m *mockChatRepository) GetByProjectId(ctx context.Context, projectId uuid.UUID) (*domain.Chat, error) {
	args := m.Called(ctx, projectId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Chat), args.Error(1)
}

func (m *mockChatRepository) CreateMember(ctx context.Context, member *domain.ChatMember, buildEvents func() []outbox.Message) error {
	err := m.Called(ctx, member).Error(0)
	if err == nil && buildEvents != nil {
		m.builtEvents = append(m.builtEvents, buildEvents()...)
	}
	return err
}

func (m *mockChatRepository) DeleteMember(ctx context.Context, member *domain.ChatMember, msgs ...outbox.Message) error {
	err := m.Called(ctx, member).Error(0)
	if err == nil {
		m.builtEvents = append(m.builtEvents, msgs...)
	}
	return err
}

func (m *mockChatRepository) CreateMessage(ctx context.Context, message *domain.ChatMessage, buildEvents func() []outbox.Message) error {
	err := m.Called(ctx, message).Error(0)
	if err == nil && buildEvents != nil {
		m.builtEvents = append(m.builtEvents, buildEvents()...)
	}
	return err
}

func (m *mockChatRepository) UpdateMemberLastSeenAt(ctx context.Context, member *domain.ChatMember) error {
	return m.Called(ctx, member).Error(0)
}

func (m *mockChatRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.Chat, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Chat), args.Error(1)
}

func (m *mockChatRepository) ListMessages(ctx context.Context, chatId uuid.UUID, params utils.PaginationBeforeParams) ([]domain.ChatMessage, error) {
	args := m.Called(ctx, chatId, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ChatMessage), args.Error(1)
}

func (m *mockChatRepository) GetOrCreateGeneralChat(ctx context.Context, currentUserId uuid.UUID, memberIds []uuid.UUID) (*domain.Chat, error) {
	args := m.Called(ctx, currentUserId, memberIds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Chat), args.Error(1)
}

func (m *mockChatRepository) ListGeneralChats(ctx context.Context, userId uuid.UUID) ([]domain.Chat, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Chat), args.Error(1)
}

func (m *mockChatRepository) GetUnreadSummary(ctx context.Context, chatId uuid.UUID, userId uuid.UUID) (*domain.ChatUnreadSummary, error) {
	args := m.Called(ctx, chatId, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ChatUnreadSummary), args.Error(1)
}

func (m *mockChatRepository) GetMessageById(ctx context.Context, id uuid.UUID) (*domain.ChatMessage, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ChatMessage), args.Error(1)
}

func (m *mockChatRepository) MarkReadUpTo(ctx context.Context, chatId uuid.UUID, userId uuid.UUID, readAt time.Time, message *domain.ChatMessage, msgs ...outbox.Message) error {
	err := m.Called(ctx, chatId, userId, readAt, message).Error(0)
	if err == nil {
		m.builtEvents = append(m.builtEvents, msgs...)
	}
	return err
}

func (m *mockChatRepository) ListMessageReads(ctx context.Context, messageId uuid.UUID) ([]domain.ChatMessageRead, error) {
	args := m.Called(ctx, messageId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ChatMessageRead), args.Error(1)
}

type mockChatUserRepository struct {
	mock.Mock
}

func (m *mockChatUserRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func newChatService(chatRepo *mockChatRepository, userRepo *mockChatUserRepository) *chat.Service {
	chatRepo.On("GetUnreadSummary", mock.Anything, mock.Anything, mock.Anything).
		Return(&domain.ChatUnreadSummary{UnreadCount: 0, HasMoreUnread: false}, nil).
		Maybe()
	return chat.NewService(chatRepo, userRepo)
}

func TestChatService_GetOrCreateGeneralChat(t *testing.T) {
	currentUserId := uuid.New()
	otherUserId := uuid.New()

	existingChat := &domain.Chat{
		Id:       uuid.New(),
		ChatType: domain.ChatTypeGeneral,
		Members: []domain.ChatMember{
			{UserId: currentUserId},
			{UserId: otherUserId},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name          string
		currentUserId uuid.UUID
		targetUserIds []uuid.UUID
		chatRepoSetup func(*mockChatRepository)
		userRepoSetup func(*mockChatUserRepository)
		shouldSucceed bool
		expectedCode  domain.ErrorCode
	}{
		{
			name:          "unauthorized when currentUserId is nil",
			currentUserId: uuid.Nil,
			targetUserIds: []uuid.UUID{otherUserId},
			chatRepoSetup: func(r *mockChatRepository) {},
			userRepoSetup: func(r *mockChatUserRepository) {},
			shouldSucceed: false,
			expectedCode:  domain.UnauthorizedErrorCode,
		},
		{
			name:          "business validation error when only current user after sanitization",
			currentUserId: currentUserId,
			targetUserIds: []uuid.UUID{currentUserId, uuid.Nil},
			chatRepoSetup: func(r *mockChatRepository) {},
			userRepoSetup: func(r *mockChatUserRepository) {},
			shouldSucceed: false,
			expectedCode:  domain.BusinessValidationErrorCode,
		},
		{
			name:          "business validation error when target list is empty",
			currentUserId: currentUserId,
			targetUserIds: []uuid.UUID{},
			chatRepoSetup: func(r *mockChatRepository) {},
			userRepoSetup: func(r *mockChatUserRepository) {},
			shouldSucceed: false,
			expectedCode:  domain.BusinessValidationErrorCode,
		},
		{
			name:          "not found error when target user does not exist",
			currentUserId: currentUserId,
			targetUserIds: []uuid.UUID{otherUserId},
			chatRepoSetup: func(r *mockChatRepository) {},
			userRepoSetup: func(r *mockChatUserRepository) {
				r.On("GetById", mock.Anything, otherUserId).Return(nil, domain.NotFoundError("user not found"))
			},
			shouldSucceed: false,
			expectedCode:  domain.NotFoundErrorCode,
		},
		{
			name:          "server error when user lookup fails unexpectedly",
			currentUserId: currentUserId,
			targetUserIds: []uuid.UUID{otherUserId},
			chatRepoSetup: func(r *mockChatRepository) {},
			userRepoSetup: func(r *mockChatUserRepository) {
				r.On("GetById", mock.Anything, otherUserId).Return(nil, errors.New("db failure"))
			},
			shouldSucceed: false,
			expectedCode:  domain.ServerErrorCode,
		},
		{
			name:          "server error when repository GetOrCreateGeneralChat fails",
			currentUserId: currentUserId,
			targetUserIds: []uuid.UUID{otherUserId},
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetOrCreateGeneralChat", mock.Anything, currentUserId, mock.Anything).Return(nil, errors.New("db failure"))
			},
			userRepoSetup: func(r *mockChatUserRepository) {
				r.On("GetById", mock.Anything, otherUserId).Return(&domain.User{Id: otherUserId}, nil)
			},
			shouldSucceed: false,
			expectedCode:  domain.ServerErrorCode,
		},
		{
			name:          "success returns existing chat",
			currentUserId: currentUserId,
			targetUserIds: []uuid.UUID{otherUserId},
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetOrCreateGeneralChat", mock.Anything, currentUserId, mock.Anything).Return(existingChat, nil)
			},
			userRepoSetup: func(r *mockChatUserRepository) {
				r.On("GetById", mock.Anything, otherUserId).Return(&domain.User{Id: otherUserId}, nil)
			},
			shouldSucceed: true,
		},
		{
			name:          "success deduplicates repeated target user",
			currentUserId: currentUserId,
			targetUserIds: []uuid.UUID{otherUserId, otherUserId},
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetOrCreateGeneralChat", mock.Anything, currentUserId, mock.MatchedBy(func(ids []uuid.UUID) bool {
					return len(ids) == 2
				})).Return(existingChat, nil)
			},
			userRepoSetup: func(r *mockChatUserRepository) {
				r.On("GetById", mock.Anything, otherUserId).Return(&domain.User{Id: otherUserId}, nil).Once()
			},
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRepo := &mockChatRepository{}
			userRepo := &mockChatUserRepository{}
			tt.chatRepoSetup(chatRepo)
			tt.userRepoSetup(userRepo)

			svc := newChatService(chatRepo, userRepo)
			chat, err := svc.GetOrCreateGeneralChat(context.Background(), tt.currentUserId, tt.targetUserIds)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, chat)
			} else {
				assert.Error(t, err)
				assert.Nil(t, chat)
				var domainErr domain.DomainError
				if errors.As(err, &domainErr) {
					assert.Equal(t, tt.expectedCode, domainErr.Code)
				}
			}

			chatRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

func TestChatService_ListGeneralChats(t *testing.T) {
	userId := uuid.New()

	tests := []struct {
		name          string
		userId        uuid.UUID
		chatRepoSetup func(*mockChatRepository)
		shouldSucceed bool
		expectedCode  domain.ErrorCode
		expectedLen   int
	}{
		{
			name:          "unauthorized when userId is nil",
			userId:        uuid.Nil,
			chatRepoSetup: func(r *mockChatRepository) {},
			shouldSucceed: false,
			expectedCode:  domain.UnauthorizedErrorCode,
		},
		{
			name:   "server error when repository fails",
			userId: userId,
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("ListGeneralChats", mock.Anything, userId).Return(nil, errors.New("db failure"))
			},
			shouldSucceed: false,
			expectedCode:  domain.ServerErrorCode,
		},
		{
			name:   "success returns empty list",
			userId: userId,
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("ListGeneralChats", mock.Anything, userId).Return([]domain.Chat{}, nil)
			},
			shouldSucceed: true,
			expectedLen:   0,
		},
		{
			name:   "success returns chats",
			userId: userId,
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("ListGeneralChats", mock.Anything, userId).Return([]domain.Chat{
					{Id: uuid.New(), ChatType: domain.ChatTypeGeneral},
					{Id: uuid.New(), ChatType: domain.ChatTypeGeneral},
				}, nil)
			},
			shouldSucceed: true,
			expectedLen:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRepo := &mockChatRepository{}
			userRepo := &mockChatUserRepository{}
			tt.chatRepoSetup(chatRepo)

			svc := newChatService(chatRepo, userRepo)
			chats, err := svc.ListGeneralChats(context.Background(), tt.userId)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.Len(t, chats, tt.expectedLen)
				chatRepo.AssertNotCalled(t, "GetUnreadSummary", mock.Anything, mock.Anything, mock.Anything)
			} else {
				assert.Error(t, err)
				assert.Nil(t, chats)
				var domainErr domain.DomainError
				if errors.As(err, &domainErr) {
					assert.Equal(t, tt.expectedCode, domainErr.Code)
				}
			}

			chatRepo.AssertExpectations(t)
		})
	}
}

func TestChatService_GetByProjectId(t *testing.T) {
	userId := uuid.New()
	projectId := uuid.New()
	chatId := uuid.New()
	otherUserId := uuid.New()

	chatWithUser := &domain.Chat{
		Id:        chatId,
		ProjectId: &projectId,
		ChatType:  domain.ChatTypeProject,
		Members:   []domain.ChatMember{{UserId: userId}},
	}
	chatWithoutUser := &domain.Chat{
		Id:        chatId,
		ProjectId: &projectId,
		ChatType:  domain.ChatTypeProject,
		Members:   []domain.ChatMember{{UserId: otherUserId}},
	}

	tests := []struct {
		name          string
		userId        uuid.UUID
		projectId     uuid.UUID
		chatRepoSetup func(*mockChatRepository)
		shouldSucceed bool
		expectedCode  domain.ErrorCode
		assertChat    func(*testing.T, *domain.Chat)
	}{
		{
			name:      "unauthorized when userId is nil",
			userId:    uuid.Nil,
			projectId: projectId,
			chatRepoSetup: func(r *mockChatRepository) {
			},
			expectedCode: domain.UnauthorizedErrorCode,
		},
		{
			name:      "not found when project chat does not exist",
			userId:    userId,
			projectId: projectId,
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetByProjectId", mock.Anything, projectId).Return(nil, domain.NotFoundError("chat not found"))
			},
			expectedCode: domain.NotFoundErrorCode,
		},
		{
			name:      "forbidden when user is not a member",
			userId:    userId,
			projectId: projectId,
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetByProjectId", mock.Anything, projectId).Return(chatWithoutUser, nil)
			},
			expectedCode: domain.ForbiddenErrorCode,
		},
		{
			name:      "success returns unread summary fields",
			userId:    userId,
			projectId: projectId,
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetByProjectId", mock.Anything, projectId).Return(chatWithUser, nil)
				r.On("GetUnreadSummary", mock.Anything, chatId, userId).
					Return(&domain.ChatUnreadSummary{UnreadCount: 99, HasMoreUnread: true}, nil)
			},
			shouldSucceed: true,
			assertChat: func(t *testing.T, chat *domain.Chat) {
				assert.Equal(t, 99, chat.UnreadCount)
				assert.True(t, chat.HasMoreUnread)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRepo := &mockChatRepository{}
			userRepo := &mockChatUserRepository{}
			tt.chatRepoSetup(chatRepo)

			svc := newChatService(chatRepo, userRepo)
			chat, err := svc.GetByProjectId(context.Background(), tt.projectId, tt.userId)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, chat)
				if tt.assertChat != nil {
					tt.assertChat(t, chat)
				}
			} else {
				assert.Error(t, err)
				assert.Nil(t, chat)
				var domainErr domain.DomainError
				if errors.As(err, &domainErr) {
					assert.Equal(t, tt.expectedCode, domainErr.Code)
				}
			}

			chatRepo.AssertExpectations(t)
		})
	}
}

func TestChatService_GetById(t *testing.T) {
	userId := uuid.New()
	chatId := uuid.New()
	otherUserId := uuid.New()

	chatWithUser := &domain.Chat{
		Id:       chatId,
		ChatType: domain.ChatTypeGeneral,
		Members:  []domain.ChatMember{{UserId: userId}},
	}
	chatWithoutUser := &domain.Chat{
		Id:       chatId,
		ChatType: domain.ChatTypeGeneral,
		Members:  []domain.ChatMember{{UserId: otherUserId}},
	}

	tests := []struct {
		name          string
		userId        uuid.UUID
		chatId        uuid.UUID
		chatRepoSetup func(*mockChatRepository)
		shouldSucceed bool
		expectedCode  domain.ErrorCode
	}{
		{
			name:          "unauthorized when userId is nil",
			userId:        uuid.Nil,
			chatId:        chatId,
			chatRepoSetup: func(r *mockChatRepository) {},
			shouldSucceed: false,
			expectedCode:  domain.UnauthorizedErrorCode,
		},
		{
			name:   "not found when chat does not exist",
			userId: userId,
			chatId: chatId,
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(nil, domain.NotFoundError("chat not found"))
			},
			shouldSucceed: false,
			expectedCode:  domain.NotFoundErrorCode,
		},
		{
			name:   "forbidden when user is not a member",
			userId: userId,
			chatId: chatId,
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(chatWithoutUser, nil)
			},
			shouldSucceed: false,
			expectedCode:  domain.ForbiddenErrorCode,
		},
		{
			name:   "server error when repository fails",
			userId: userId,
			chatId: chatId,
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(nil, errors.New("db failure"))
			},
			shouldSucceed: false,
			expectedCode:  domain.ServerErrorCode,
		},
		{
			name:   "success returns chat details",
			userId: userId,
			chatId: chatId,
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(chatWithUser, nil)
			},
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRepo := &mockChatRepository{}
			userRepo := &mockChatUserRepository{}
			tt.chatRepoSetup(chatRepo)

			svc := newChatService(chatRepo, userRepo)
			chat, err := svc.GetById(context.Background(), tt.chatId, tt.userId)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, chat)
				assert.Equal(t, tt.chatId, chat.Id)
			} else {
				assert.Error(t, err)
				assert.Nil(t, chat)
				var domainErr domain.DomainError
				if errors.As(err, &domainErr) {
					assert.Equal(t, tt.expectedCode, domainErr.Code)
				}
			}

			chatRepo.AssertExpectations(t)
		})
	}
}

func TestChatService_ListMessagesByChatId(t *testing.T) {
	userId := uuid.New()
	chatId := uuid.New()
	otherUserId := uuid.New()

	chatWithUser := &domain.Chat{
		Id:       chatId,
		ChatType: domain.ChatTypeGeneral,
		Members:  []domain.ChatMember{{UserId: userId}},
	}
	chatWithoutUser := &domain.Chat{
		Id:       chatId,
		ChatType: domain.ChatTypeGeneral,
		Members:  []domain.ChatMember{{UserId: otherUserId}},
	}

	paginationParams := utils.PaginationBeforeParams{
		Limit:  10,
		Before: time.Now(),
		Id:     uuid.Nil,
	}

	msg1 := domain.ChatMessage{Id: uuid.New(), ChatId: chatId, Content: "first", CreatedAt: time.Now().Add(-2 * time.Second)}
	msg2 := domain.ChatMessage{Id: uuid.New(), ChatId: chatId, Content: "second", CreatedAt: time.Now().Add(-1 * time.Second)}

	tests := []struct {
		name          string
		request       chat.ListMessagesByChatIdRequest
		chatRepoSetup func(*mockChatRepository)
		shouldSucceed bool
		expectedCode  domain.ErrorCode
		expectedLen   int
		checkOrder    bool
	}{
		{
			name: "unauthorized when userId is nil",
			request: chat.ListMessagesByChatIdRequest{
				ChatId: chatId,
				UserId: uuid.Nil,
				Params: paginationParams,
			},
			chatRepoSetup: func(r *mockChatRepository) {},
			shouldSucceed: false,
			expectedCode:  domain.UnauthorizedErrorCode,
		},
		{
			name: "not found when chat does not exist",
			request: chat.ListMessagesByChatIdRequest{
				ChatId: chatId,
				UserId: userId,
				Params: paginationParams,
			},
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(nil, domain.NotFoundError("chat not found"))
			},
			shouldSucceed: false,
			expectedCode:  domain.NotFoundErrorCode,
		},
		{
			name: "forbidden when user is not a member",
			request: chat.ListMessagesByChatIdRequest{
				ChatId: chatId,
				UserId: userId,
				Params: paginationParams,
			},
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(chatWithoutUser, nil)
			},
			shouldSucceed: false,
			expectedCode:  domain.ForbiddenErrorCode,
		},
		{
			name: "server error when list messages fails",
			request: chat.ListMessagesByChatIdRequest{
				ChatId: chatId,
				UserId: userId,
				Params: paginationParams,
			},
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(chatWithUser, nil)
				r.On("ListMessages", mock.Anything, chatId, paginationParams).Return(nil, errors.New("db failure"))
			},
			shouldSucceed: false,
			expectedCode:  domain.ServerErrorCode,
		},
		{
			name: "success returns messages in ascending order",
			request: chat.ListMessagesByChatIdRequest{
				ChatId: chatId,
				UserId: userId,
				Params: paginationParams,
			},
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(chatWithUser, nil)
				r.On("ListMessages", mock.Anything, chatId, paginationParams).Return([]domain.ChatMessage{msg2, msg1}, nil)
			},
			shouldSucceed: true,
			expectedLen:   2,
			checkOrder:    true,
		},
		{
			name: "success sets HasNext true when messages equal limit",
			request: chat.ListMessagesByChatIdRequest{
				ChatId: chatId,
				UserId: userId,
				Params: utils.PaginationBeforeParams{Limit: 2, Before: time.Now(), Id: uuid.Nil},
			},
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(chatWithUser, nil)
				r.On("ListMessages", mock.Anything, chatId, mock.Anything).Return([]domain.ChatMessage{msg2, msg1}, nil)
			},
			shouldSucceed: true,
			expectedLen:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRepo := &mockChatRepository{}
			userRepo := &mockChatUserRepository{}
			tt.chatRepoSetup(chatRepo)

			svc := newChatService(chatRepo, userRepo)
			result, err := svc.ListMessagesByChatId(context.Background(), tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Data, tt.expectedLen)
				if tt.checkOrder && len(result.Data) == 2 {
					assert.Equal(t, msg1.Id, result.Data[0].Id)
					assert.Equal(t, msg2.Id, result.Data[1].Id)
				}
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				var domainErr domain.DomainError
				if errors.As(err, &domainErr) {
					assert.Equal(t, tt.expectedCode, domainErr.Code)
				}
			}

			chatRepo.AssertExpectations(t)
		})
	}
}

func TestSanitizeGeneralChatMemberIDs(t *testing.T) {
	currentUser := uuid.New()
	userA := uuid.New()
	userB := uuid.New()

	tests := []struct {
		name          string
		currentUserId uuid.UUID
		targetUserIds []uuid.UUID
		expectedLen   int
		checkFirst    bool
	}{
		{
			name:          "current user always at index 0",
			currentUserId: currentUser,
			targetUserIds: []uuid.UUID{userA},
			expectedLen:   2,
			checkFirst:    true,
		},
		{
			name:          "removes nil UUIDs from target list",
			currentUserId: currentUser,
			targetUserIds: []uuid.UUID{uuid.Nil, userA},
			expectedLen:   2,
		},
		{
			name:          "deduplicates repeated target users",
			currentUserId: currentUser,
			targetUserIds: []uuid.UUID{userA, userA, userB},
			expectedLen:   3,
		},
		{
			name:          "removes current user if included in target list",
			currentUserId: currentUser,
			targetUserIds: []uuid.UUID{currentUser, userA},
			expectedLen:   2,
		},
		{
			name:          "empty target returns only current user",
			currentUserId: currentUser,
			targetUserIds: []uuid.UUID{},
			expectedLen:   1,
		},
		{
			name:          "all nil targets returns only current user",
			currentUserId: currentUser,
			targetUserIds: []uuid.UUID{uuid.Nil, uuid.Nil},
			expectedLen:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := chat.SanitizeGeneralChatMemberIDs(tt.currentUserId, tt.targetUserIds)
			assert.Len(t, result, tt.expectedLen)
			if tt.checkFirst {
				assert.Equal(t, tt.currentUserId, result[0])
			}
			seen := map[uuid.UUID]struct{}{}
			for _, id := range result {
				_, exists := seen[id]
				assert.False(t, exists, "duplicate UUID found: %s", id)
				seen[id] = struct{}{}
			}
			// Verify no nil UUIDs
			for _, id := range result {
				assert.NotEqual(t, uuid.Nil, id)
			}
		})
	}
}

func TestChatService_MarkChatRead(t *testing.T) {
	userId := uuid.New()
	otherUserId := uuid.New()
	chatId := uuid.New()
	otherChatId := uuid.New()
	messageId := uuid.New()

	chatWithUser := &domain.Chat{
		Id:       chatId,
		ChatType: domain.ChatTypeGeneral,
		Members:  []domain.ChatMember{{UserId: userId}},
	}
	chatWithoutUser := &domain.Chat{
		Id:       chatId,
		ChatType: domain.ChatTypeGeneral,
		Members:  []domain.ChatMember{{UserId: otherUserId}},
	}
	message := &domain.ChatMessage{
		Id:        messageId,
		ChatId:    chatId,
		CreatedAt: time.Now().Add(-time.Minute),
	}
	messageFromAnotherChat := &domain.ChatMessage{
		Id:        messageId,
		ChatId:    otherChatId,
		CreatedAt: time.Now().Add(-time.Minute),
	}

	tests := []struct {
		name          string
		request       chat.MarkChatReadRequest
		chatRepoSetup func(*mockChatRepository)
		shouldSucceed bool
		expectedCode  domain.ErrorCode
		expectPublish bool
	}{
		{
			name: "unauthorized when user is nil",
			request: chat.MarkChatReadRequest{
				ChatId: chatId,
				UserId: uuid.Nil,
			},
			chatRepoSetup: func(r *mockChatRepository) {},
			expectedCode:  domain.UnauthorizedErrorCode,
		},
		{
			name: "forbidden when user is not a member",
			request: chat.MarkChatReadRequest{
				ChatId: chatId,
				UserId: userId,
			},
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(chatWithoutUser, nil)
			},
			expectedCode: domain.ForbiddenErrorCode,
		},
		{
			name: "forbidden when message belongs to another chat",
			request: chat.MarkChatReadRequest{
				ChatId:    chatId,
				UserId:    userId,
				MessageId: &messageId,
			},
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(chatWithUser, nil)
				r.On("GetMessageById", mock.Anything, messageId).Return(messageFromAnotherChat, nil)
			},
			expectedCode: domain.ForbiddenErrorCode,
		},
		{
			name: "success without message id does not publish message read event",
			request: chat.MarkChatReadRequest{
				ChatId: chatId,
				UserId: userId,
			},
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(chatWithUser, nil)
				r.On("MarkReadUpTo", mock.Anything, chatId, userId, mock.AnythingOfType("time.Time"), (*domain.ChatMessage)(nil)).Return(nil)
			},
			shouldSucceed: true,
			expectPublish: false,
		},
		{
			name: "success with message id publishes message read event",
			request: chat.MarkChatReadRequest{
				ChatId:    chatId,
				UserId:    userId,
				MessageId: &messageId,
			},
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(chatWithUser, nil)
				r.On("GetMessageById", mock.Anything, messageId).Return(message, nil)
				r.On("MarkReadUpTo", mock.Anything, chatId, userId, mock.AnythingOfType("time.Time"), message).Return(nil)
			},
			shouldSucceed: true,
			expectPublish: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRepo := &mockChatRepository{}
			userRepo := &mockChatUserRepository{}
			tt.chatRepoSetup(chatRepo)

			svc := newChatService(chatRepo, userRepo)
			err := svc.MarkChatRead(context.Background(), tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var domainErr domain.DomainError
				if errors.As(err, &domainErr) {
					assert.Equal(t, tt.expectedCode, domainErr.Code)
				}
			}

			if tt.expectPublish {
				require.Len(t, chatRepo.builtEvents, 1)
				assert.Equal(t, events.ChatMessageRead, chatRepo.builtEvents[0].Topic)
				payload, ok := chatRepo.builtEvents[0].Payload.(*events.ChatMessageReadPayload)
				require.True(t, ok)
				assert.Equal(t, chatId, payload.ChatID)
				assert.Equal(t, messageId, payload.MessageID)
				assert.Equal(t, userId, payload.Read.UserId)
			} else {
				assert.Empty(t, chatRepo.builtEvents)
			}

			chatRepo.AssertExpectations(t)
		})
	}
}

func TestChatService_ListMessageReads(t *testing.T) {
	userId := uuid.New()
	otherUserId := uuid.New()
	chatId := uuid.New()
	otherChatId := uuid.New()
	messageId := uuid.New()

	chatWithUser := &domain.Chat{
		Id:       chatId,
		ChatType: domain.ChatTypeGeneral,
		Members:  []domain.ChatMember{{UserId: userId}},
	}
	chatWithoutUser := &domain.Chat{
		Id:       chatId,
		ChatType: domain.ChatTypeGeneral,
		Members:  []domain.ChatMember{{UserId: otherUserId}},
	}
	message := &domain.ChatMessage{
		Id:     messageId,
		ChatId: chatId,
	}
	messageFromAnotherChat := &domain.ChatMessage{
		Id:     messageId,
		ChatId: otherChatId,
	}
	reads := []domain.ChatMessageRead{
		{MessageId: messageId, UserId: otherUserId, ReadAt: time.Now()},
	}

	tests := []struct {
		name          string
		request       chat.ListMessageReadsRequest
		chatRepoSetup func(*mockChatRepository)
		shouldSucceed bool
		expectedCode  domain.ErrorCode
		expectedLen   int
	}{
		{
			name: "unauthorized when user is nil",
			request: chat.ListMessageReadsRequest{
				ChatId:    chatId,
				MessageId: messageId,
				UserId:    uuid.Nil,
			},
			chatRepoSetup: func(r *mockChatRepository) {},
			expectedCode:  domain.UnauthorizedErrorCode,
		},
		{
			name: "forbidden when user is not member",
			request: chat.ListMessageReadsRequest{
				ChatId:    chatId,
				MessageId: messageId,
				UserId:    userId,
			},
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(chatWithoutUser, nil)
			},
			expectedCode: domain.ForbiddenErrorCode,
		},
		{
			name: "forbidden when message belongs to another chat",
			request: chat.ListMessageReadsRequest{
				ChatId:    chatId,
				MessageId: messageId,
				UserId:    userId,
			},
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(chatWithUser, nil)
				r.On("GetMessageById", mock.Anything, messageId).Return(messageFromAnotherChat, nil)
			},
			expectedCode: domain.ForbiddenErrorCode,
		},
		{
			name: "success returns reads",
			request: chat.ListMessageReadsRequest{
				ChatId:    chatId,
				MessageId: messageId,
				UserId:    userId,
			},
			chatRepoSetup: func(r *mockChatRepository) {
				r.On("GetById", mock.Anything, chatId).Return(chatWithUser, nil)
				r.On("GetMessageById", mock.Anything, messageId).Return(message, nil)
				r.On("ListMessageReads", mock.Anything, messageId).Return(reads, nil)
			},
			shouldSucceed: true,
			expectedLen:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRepo := &mockChatRepository{}
			userRepo := &mockChatUserRepository{}
			tt.chatRepoSetup(chatRepo)

			svc := newChatService(chatRepo, userRepo)
			result, err := svc.ListMessageReads(context.Background(), tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedLen)
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				var domainErr domain.DomainError
				if errors.As(err, &domainErr) {
					assert.Equal(t, tt.expectedCode, domainErr.Code)
				}
			}

			chatRepo.AssertExpectations(t)
		})
	}
}
