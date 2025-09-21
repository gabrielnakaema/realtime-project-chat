package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type chatService interface {
	GetByProjectId(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Chat, error)
	CreateMessage(ctx context.Context, request service.CreateChatMessageRequest) (*domain.ChatMessage, error)
	ListMessagesByProjectId(ctx context.Context, request service.ListMessagesByProjectIdRequest) (*utils.CursorPaginated[domain.ChatMessage], error)
}

type ChatHandler struct {
	chatService chatService
}

func NewChatHandler(chatService chatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

func (ch *ChatHandler) GetChatByProjectId(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "id")
	if projectId == "" {
		BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	parsedProjectId, err := uuid.Parse(projectId)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	userId := UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		UnauthorizedResponse(w, "unauthorized")
		return
	}

	chat, err := ch.chatService.GetByProjectId(r.Context(), parsedProjectId, userId)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, chat, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (ch *ChatHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var request CreateMessageRequest
	err := utils.ReadJSON(w, r, &request)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	request.Validate(v)
	if !v.Valid() {
		ValidationFailedResponse(w, v)
		return
	}

	userId := UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		UnauthorizedResponse(w, "unauthorized")
		return
	}

	serviceRequest := service.CreateChatMessageRequest{
		ChatId:  request.ChatId,
		UserId:  userId,
		Content: request.Content,
	}

	message, err := ch.chatService.CreateMessage(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, message, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (ch *ChatHandler) ListMessagesByProjectId(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "id")
	if projectId == "" {
		BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	limit := utils.GetQueryInt(r, "limit", 10)
	if limit <= 0 {
		BadRequestResponse(w, errors.New("limit must be greater than 0"))
		return
	}

	if limit > 50 {
		BadRequestResponse(w, errors.New("limit must be less than 50"))
		return
	}

	before := utils.GetQueryString(r, "before", "")
	beforeTime := time.Now()
	if before != "" {
		date, err := time.Parse(time.RFC3339, before)
		if err != nil {
			BadRequestResponse(w, errors.New("invalid before date"))
			return
		}
		beforeTime = date
	}

	beforeId := utils.GetQueryString(r, "id", "")
	beforeIdUUID := uuid.Nil
	if beforeId != "" {
		parsedBeforeId, err := uuid.Parse(beforeId)
		if err != nil {
			BadRequestResponse(w, err)
			return
		}
		beforeIdUUID = parsedBeforeId
	}

	parsedProjectId, err := uuid.Parse(projectId)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	userId := UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		UnauthorizedResponse(w, "unauthorized")
		return
	}

	paginationParams := utils.PaginationBeforeParams{
		Limit:  limit,
		Before: beforeTime,
		Id:     beforeIdUUID,
	}

	serviceRequest := service.ListMessagesByProjectIdRequest{
		ProjectId: parsedProjectId,
		UserId:    userId,
		Params:    paginationParams,
	}

	messages, err := ch.chatService.ListMessagesByProjectId(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, messages, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}
