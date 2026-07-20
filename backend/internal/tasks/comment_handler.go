package tasks

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/auth"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/handlers"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type taskCommentService interface {
	Create(ctx context.Context, request CreateTaskCommentRequest) (*domain.TaskComment, error)
	ListByTaskID(ctx context.Context, request ListTaskCommentsRequest) (*utils.CursorPaginated[domain.TaskComment], error)
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
		handlers.BadRequestResponse(w, err)
		return
	}

	var request CreateTaskCommentBody
	err = utils.ReadJSON(w, r, &request)
	if err != nil {
		handlers.BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	request.Validate(v)
	if !v.Valid() {
		handlers.ValidationFailedResponse(w, v)
		return
	}

	userID := auth.UserIdFromContext(r.Context())

	comment, err := h.taskCommentService.Create(r.Context(), CreateTaskCommentRequest{
		TaskID:          parsedTaskID,
		RequestUserID:   userID,
		Content:         request.Content,
		ParentCommentID: request.ParentCommentID,
	})
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, comment, nil)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskCommentHandler) ListByTaskID(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	parsedTaskID, err := uuid.Parse(taskID)
	if err != nil {
		handlers.BadRequestResponse(w, err)
		return
	}

	limit := utils.GetQueryInt(r, "limit", 10)
	if limit <= 0 {
		handlers.BadRequestResponse(w, errors.New("limit must be greater than 0"))
		return
	}

	if limit > 50 {
		handlers.BadRequestResponse(w, errors.New("limit must be less than 50"))
		return
	}

	before := utils.GetQueryString(r, "before", "")
	after := utils.GetQueryString(r, "after", "")
	if before != "" && after != "" {
		handlers.BadRequestResponse(w, errors.New("before and after cannot be used together"))
		return
	}

	var beforeTime *time.Time
	if before != "" {
		parsedBefore, err := time.Parse(time.RFC3339, before)
		if err != nil {
			handlers.BadRequestResponse(w, errors.New("invalid before date"))
			return
		}
		beforeTime = &parsedBefore
	}

	var afterTime *time.Time
	if after != "" {
		parsedAfter, err := time.Parse(time.RFC3339, after)
		if err != nil {
			handlers.BadRequestResponse(w, errors.New("invalid after date"))
			return
		}
		afterTime = &parsedAfter
	}

	if beforeTime == nil && afterTime == nil {
		now := time.Now()
		beforeTime = &now
	}

	beforeCommentID := utils.GetQueryString(r, "comment_id", "")
	var parsedBeforeCommentID *uuid.UUID
	if beforeCommentID != "" {
		id, err := uuid.Parse(beforeCommentID)
		if err != nil {
			handlers.BadRequestResponse(w, err)
			return
		}
		parsedBeforeCommentID = &id
	}

	afterCommentID := utils.GetQueryString(r, "after_comment_id", "")
	var parsedAfterCommentID *uuid.UUID
	if afterCommentID != "" {
		id, err := uuid.Parse(afterCommentID)
		if err != nil {
			handlers.BadRequestResponse(w, err)
			return
		}
		parsedAfterCommentID = &id
	}

	userID := auth.UserIdFromContext(r.Context())

	comments, err := h.taskCommentService.ListByTaskID(r.Context(), ListTaskCommentsRequest{
		TaskID:          parsedTaskID,
		RequestUserID:   userID,
		Limit:           int(limit),
		Before:          beforeTime,
		BeforeCommentID: parsedBeforeCommentID,
		After:           afterTime,
		AfterCommentID:  parsedAfterCommentID,
	})
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, comments, nil)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}
}
