package notification

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

type notificationService interface {
	List(ctx context.Context, request ListRequest) (*utils.CursorPaginated[domain.Notification], error)
	CountUnread(ctx context.Context, userId uuid.UUID) (int, error)
	MarkRead(ctx context.Context, request MarkReadRequest) error
	MarkAllRead(ctx context.Context, userId uuid.UUID) error
}

type Handler struct {
	notificationService notificationService
}

func NewHandler(notificationService notificationService) *Handler {
	return &Handler{
		notificationService: notificationService,
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	limit := platformhttp.GetQueryInt(r, "limit", 10)
	before := platformhttp.GetQueryString(r, "before", "")
	beforeId := platformhttp.GetQueryString(r, "id", "")

	v := validator.New()
	v.Check("limit", "limit must be greater than 0", limit > 0)
	v.Check("limit", "limit must be less than or equal to 50", limit <= 50)

	beforeTime := time.Now()
	if before != "" {
		date, err := time.Parse(time.RFC3339, before)
		if err != nil {
			v.Add("before", "invalid before date")
		} else {
			beforeTime = date
		}
	}

	beforeIdUUID := uuid.Nil
	if beforeId != "" {
		parsedBeforeId, err := uuid.Parse(beforeId)
		if err != nil {
			v.Add("id", "invalid before id")
		} else {
			beforeIdUUID = parsedBeforeId
		}
	}

	if !v.Valid() {
		httperr.ValidationFailedResponse(w, v)
		return
	}

	notifications, err := h.notificationService.List(r.Context(), ListRequest{
		UserId:          userId,
		BeforeCreatedAt: beforeTime,
		BeforeId:        beforeIdUUID,
		Limit:           limit,
	})
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	if err := platformhttp.WriteJSON(w, http.StatusOK, notifications, nil); err != nil {
		httperr.ErrorResponse(w, r, err)
	}
}

func (h *Handler) CountUnread(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	count, err := h.notificationService.CountUnread(r.Context(), userId)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	if err := platformhttp.WriteJSON(w, http.StatusOK, map[string]int{"count": count}, nil); err != nil {
		httperr.ErrorResponse(w, r, err)
	}
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	id := chi.URLParam(r, "id")
	notificationId, err := uuid.Parse(id)
	if err != nil {
		httperr.BadRequestResponse(w, errors.New("invalid notification id"))
		return
	}

	err = h.notificationService.MarkRead(r.Context(), MarkReadRequest{
		NotificationId: notificationId,
		UserId:         userId,
	})
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	if err := platformhttp.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true}, nil); err != nil {
		httperr.ErrorResponse(w, r, err)
	}
}

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	if err := h.notificationService.MarkAllRead(r.Context(), userId); err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	if err := platformhttp.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true}, nil); err != nil {
		httperr.ErrorResponse(w, r, err)
	}
}
