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
	"github.com/go-chi/chi/v5"
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
	projectId := chi.URLParam(r, "id")
	if projectId == "" {
		httperr.BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	parsedProjectId, err := uuid.Parse(projectId)
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
	chatId := chi.URLParam(r, "chatId")
	if chatId == "" {
		httperr.BadRequestResponse(w, errors.New("chatId is required"))
		return
	}

	parsedChatId, err := uuid.Parse(chatId)
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
	chatId := chi.URLParam(r, "chatId")
	if chatId == "" {
		httperr.BadRequestResponse(w, errors.New("chatId is required"))
		return
	}

	limit := platformhttp.GetQueryInt(r, "limit", 10)
	if limit <= 0 {
		httperr.BadRequestResponse(w, errors.New("limit must be greater than 0"))
		return
	}

	if limit > 50 {
		httperr.BadRequestResponse(w, errors.New("limit must be less than 50"))
		return
	}

	before := platformhttp.GetQueryString(r, "before", "")
	beforeTime := time.Now()
	if before != "" {
		date, err := time.Parse(time.RFC3339, before)
		if err != nil {
			httperr.BadRequestResponse(w, errors.New("invalid before date"))
			return
		}
		beforeTime = date
	}

	beforeId := platformhttp.GetQueryString(r, "id", "")
	beforeIdUUID := uuid.Nil
	if beforeId != "" {
		parsedBeforeId, err := uuid.Parse(beforeId)
		if err != nil {
			httperr.BadRequestResponse(w, err)
			return
		}
		beforeIdUUID = parsedBeforeId
	}

	parsedChatId, err := uuid.Parse(chatId)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
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
	chatId := chi.URLParam(r, "chatId")
	if chatId == "" {
		httperr.BadRequestResponse(w, errors.New("chatId is required"))
		return
	}

	parsedChatId, err := uuid.Parse(chatId)
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
	chatId := chi.URLParam(r, "chatId")
	messageId := chi.URLParam(r, "messageId")
	if chatId == "" || messageId == "" {
		httperr.BadRequestResponse(w, errors.New("chatId and messageId are required"))
		return
	}

	parsedChatId, err := uuid.Parse(chatId)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	parsedMessageId, err := uuid.Parse(messageId)
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
	projectId := chi.URLParam(r, "id")
	if projectId == "" {
		httperr.BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	limit := platformhttp.GetQueryInt(r, "limit", 10)
	if limit <= 0 {
		httperr.BadRequestResponse(w, errors.New("limit must be greater than 0"))
		return
	}

	if limit > 50 {
		httperr.BadRequestResponse(w, errors.New("limit must be less than 50"))
		return
	}

	before := platformhttp.GetQueryString(r, "before", "")
	beforeTime := time.Now()
	if before != "" {
		date, err := time.Parse(time.RFC3339, before)
		if err != nil {
			httperr.BadRequestResponse(w, errors.New("invalid before date"))
			return
		}
		beforeTime = date
	}

	beforeId := platformhttp.GetQueryString(r, "id", "")
	beforeIdUUID := uuid.Nil
	if beforeId != "" {
		parsedBeforeId, err := uuid.Parse(beforeId)
		if err != nil {
			httperr.BadRequestResponse(w, err)
			return
		}
		beforeIdUUID = parsedBeforeId
	}

	parsedProjectId, err := uuid.Parse(projectId)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
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
