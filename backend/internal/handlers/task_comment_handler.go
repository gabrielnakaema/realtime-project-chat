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

type taskCommentService interface {
	Create(ctx context.Context, request service.CreateTaskCommentRequest) (*domain.TaskComment, error)
	ListByTaskID(ctx context.Context, request service.ListTaskCommentsRequest) (*utils.CursorPaginated[domain.TaskComment], error)
}

type TaskCommentHandler struct {
	taskCommentService taskCommentService
}

func NewTaskCommentHandler(taskCommentService taskCommentService) *TaskCommentHandler {
	return &TaskCommentHandler{
		taskCommentService: taskCommentService,
	}
}

func (h *TaskCommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	parsedTaskID, err := uuid.Parse(taskID)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	var request CreateTaskCommentRequest
	err = utils.ReadJSON(w, r, &request)
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

	userID := UserIdFromContext(r.Context())

	comment, err := h.taskCommentService.Create(r.Context(), service.CreateTaskCommentRequest{
		TaskID:          parsedTaskID,
		RequestUserID:   userID,
		Content:         request.Content,
		ParentCommentID: request.ParentCommentID,
	})
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, comment, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskCommentHandler) ListByTaskID(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	parsedTaskID, err := uuid.Parse(taskID)
	if err != nil {
		BadRequestResponse(w, err)
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
		parsedBefore, err := time.Parse(time.RFC3339, before)
		if err != nil {
			BadRequestResponse(w, errors.New("invalid before date"))
			return
		}
		beforeTime = parsedBefore
	}

	beforeCommentID := utils.GetQueryString(r, "comment_id", "")
	parsedBeforeCommentID := uuid.Nil
	if beforeCommentID != "" {
		parsedBeforeCommentID, err = uuid.Parse(beforeCommentID)
		if err != nil {
			BadRequestResponse(w, err)
			return
		}
	}

	userID := UserIdFromContext(r.Context())

	comments, err := h.taskCommentService.ListByTaskID(r.Context(), service.ListTaskCommentsRequest{
		TaskID:          parsedTaskID,
		RequestUserID:   userID,
		Limit:           int(limit),
		Before:          beforeTime,
		BeforeCommentID: parsedBeforeCommentID,
	})
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, comments, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}
