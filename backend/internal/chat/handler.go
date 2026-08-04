package chat

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/auth"
	platformhttp "github.com/gabrielnakaema/project-chat/internal/platform/http"
	"github.com/gabrielnakaema/project-chat/internal/platform/httperr"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/google/uuid"
)

type chatService interface {
	GetByProjectId(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Chat, error)
	GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Chat, error)
	CreateMessage(ctx context.Context, request CreateChatMessageRequest) (*domain.ChatMessage, error)
	ListMessagesByProjectId(ctx context.Context, request ListMessagesByProjectIdRequest) (*utils.CursorPaginated[domain.ChatMessage], error)
	GetOrCreateGeneralChat(ctx context.Context, currentUserId uuid.UUID, targetUserIds []uuid.UUID) (*domain.Chat, error)
	ListGeneralChats(ctx context.Context, userId uuid.UUID) ([]domain.Chat, error)
	ListMessagesByChatId(ctx context.Context, request ListMessagesByChatIdRequest) (*utils.CursorPaginated[domain.ChatMessage], error)
	MarkChatRead(ctx context.Context, request MarkChatReadRequest) error
	ListMessageReads(ctx context.Context, request ListMessageReadsRequest) ([]domain.ChatMessageRead, error)
}

type Handler struct {
	chatService chatService
}

func NewHandler(chatService chatService) *Handler {
	return &Handler{
		chatService: chatService,
	}
}

func (ch *Handler) GetChatByProjectId(w http.ResponseWriter, r *http.Request) {
	parsedProjectId, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	userId := auth.UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		httperr.UnauthorizedResponse(w, "unauthorized")
		return
	}

	chat, err := ch.chatService.GetByProjectId(r.Context(), parsedProjectId, userId)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, chat, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (ch *Handler) GetChatById(w http.ResponseWriter, r *http.Request) {
	parsedChatId, err := platformhttp.ParseURLUUID(r, "chatId")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	userId := auth.UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		httperr.UnauthorizedResponse(w, "unauthorized")
		return
	}

	chat, err := ch.chatService.GetById(r.Context(), parsedChatId, userId)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, chat, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (ch *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var request CreateMessageRequest
	err := platformhttp.ReadJSON(w, r, &request)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	request.Validate(v)
	if !v.Valid() {
		httperr.ValidationFailedResponse(w, v)
		return
	}

	userId := auth.UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		httperr.UnauthorizedResponse(w, "unauthorized")
		return
	}

	serviceRequest := CreateChatMessageRequest{
		ChatId:  request.ChatId,
		UserId:  userId,
		Content: request.Content,
	}

	message, err := ch.chatService.CreateMessage(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusCreated, message, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (ch *Handler) GetOrCreateGeneralChat(w http.ResponseWriter, r *http.Request) {
	var request GetOrCreateGeneralChatRequest
	err := platformhttp.ReadJSON(w, r, &request)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	request.Validate(v)
	if !v.Valid() {
		httperr.ValidationFailedResponse(w, v)
		return
	}

	userId := auth.UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		httperr.UnauthorizedResponse(w, "unauthorized")
		return
	}

	chat, err := ch.chatService.GetOrCreateGeneralChat(r.Context(), userId, request.UserIds)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, chat, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (ch *Handler) ListGeneralChats(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		httperr.UnauthorizedResponse(w, "unauthorized")
		return
	}

	chats, err := ch.chatService.ListGeneralChats(r.Context(), userId)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, chats, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (ch *Handler) ListChatMessages(w http.ResponseWriter, r *http.Request) {
	parsedChatId, err := platformhttp.ParseURLUUID(r, "chatId")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	limit, err := platformhttp.ParseLimit(r, 10, 50)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	beforeTime := time.Now()
	parsedBeforeTime, err := platformhttp.ParseRFC3339Cursor(r, "before")
	if err != nil {
		httperr.BadRequestResponse(w, errors.New("invalid before date"))
		return
	}
	if parsedBeforeTime != nil {
		beforeTime = *parsedBeforeTime
	}

	beforeIdUUID := uuid.Nil
	parsedBeforeId, err := platformhttp.ParseOptionalQueryUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}
	if parsedBeforeId != nil {
		beforeIdUUID = *parsedBeforeId
	}

	userId := auth.UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		httperr.UnauthorizedResponse(w, "unauthorized")
		return
	}

	paginationParams := utils.PaginationBeforeParams{
		Limit:  limit,
		Before: beforeTime,
		Id:     beforeIdUUID,
	}

	serviceRequest := ListMessagesByChatIdRequest{
		ChatId: parsedChatId,
		UserId: userId,
		Params: paginationParams,
	}

	messages, err := ch.chatService.ListMessagesByChatId(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, messages, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (ch *Handler) MarkChatRead(w http.ResponseWriter, r *http.Request) {
	parsedChatId, err := platformhttp.ParseURLUUID(r, "chatId")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	var request MarkChatReadBody
	err = platformhttp.ReadJSON(w, r, &request)
	if err != nil && !errors.Is(err, http.ErrBodyNotAllowed) && err.Error() != "EOF" {
		httperr.BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	request.Validate(v)
	if !v.Valid() {
		httperr.ValidationFailedResponse(w, v)
		return
	}

	userId := auth.UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		httperr.UnauthorizedResponse(w, "unauthorized")
		return
	}

	err = ch.chatService.MarkChatRead(r.Context(), MarkChatReadRequest{
		ChatId:    parsedChatId,
		UserId:    userId,
		MessageId: request.MessageId,
	})
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ch *Handler) ListMessageReads(w http.ResponseWriter, r *http.Request) {
	parsedChatId, err := platformhttp.ParseURLUUID(r, "chatId")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	parsedMessageId, err := platformhttp.ParseURLUUID(r, "messageId")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	userId := auth.UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		httperr.UnauthorizedResponse(w, "unauthorized")
		return
	}

	reads, err := ch.chatService.ListMessageReads(r.Context(), ListMessageReadsRequest{
		ChatId:    parsedChatId,
		MessageId: parsedMessageId,
		UserId:    userId,
	})
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, reads, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (ch *Handler) ListMessagesByProjectId(w http.ResponseWriter, r *http.Request) {
	parsedProjectId, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	limit, err := platformhttp.ParseLimit(r, 10, 50)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	beforeTime := time.Now()
	parsedBeforeTime, err := platformhttp.ParseRFC3339Cursor(r, "before")
	if err != nil {
		httperr.BadRequestResponse(w, errors.New("invalid before date"))
		return
	}
	if parsedBeforeTime != nil {
		beforeTime = *parsedBeforeTime
	}

	beforeIdUUID := uuid.Nil
	parsedBeforeId, err := platformhttp.ParseOptionalQueryUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}
	if parsedBeforeId != nil {
		beforeIdUUID = *parsedBeforeId
	}

	userId := auth.UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		httperr.UnauthorizedResponse(w, "unauthorized")
		return
	}

	paginationParams := utils.PaginationBeforeParams{
		Limit:  limit,
		Before: beforeTime,
		Id:     beforeIdUUID,
	}

	serviceRequest := ListMessagesByProjectIdRequest{
		ProjectId: parsedProjectId,
		UserId:    userId,
		Params:    paginationParams,
	}

	messages, err := ch.chatService.ListMessagesByProjectId(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, messages, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}
