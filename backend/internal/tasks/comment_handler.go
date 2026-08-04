package tasks

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
	parsedTaskID, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	var request CreateTaskCommentBody
	err = platformhttp.ReadJSON(w, r, &request)
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

	userID := auth.UserIdFromContext(r.Context())

	comment, err := h.taskCommentService.Create(r.Context(), CreateTaskCommentRequest{
		TaskID:          parsedTaskID,
		RequestUserID:   userID,
		Content:         request.Content,
		ParentCommentID: request.ParentCommentID,
	})
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusCreated, comment, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskCommentHandler) ListByTaskID(w http.ResponseWriter, r *http.Request) {
	parsedTaskID, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	limit, err := platformhttp.ParseLimit(r, 10, 50)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	beforeTime, err := platformhttp.ParseRFC3339Cursor(r, "before")
	if err != nil {
		httperr.BadRequestResponse(w, errors.New("invalid before date"))
		return
	}
	afterTime, err := platformhttp.ParseRFC3339Cursor(r, "after")
	if err != nil {
		httperr.BadRequestResponse(w, errors.New("invalid after date"))
		return
	}
	if beforeTime != nil && afterTime != nil {
		httperr.BadRequestResponse(w, errors.New("before and after cannot be used together"))
		return
	}

	if beforeTime == nil && afterTime == nil {
		now := time.Now()
		beforeTime = &now
	}

	parsedBeforeCommentID, err := platformhttp.ParseOptionalQueryUUID(r, "comment_id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	parsedAfterCommentID, err := platformhttp.ParseOptionalQueryUUID(r, "after_comment_id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
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
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, comments, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}
