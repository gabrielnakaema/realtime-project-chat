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

type notificationService interface {
	List(ctx context.Context, request service.ListNotificationsRequest) (*utils.CursorPaginated[domain.Notification], error)
	CountUnread(ctx context.Context, userId uuid.UUID) (int, error)
	MarkRead(ctx context.Context, request service.MarkNotificationReadRequest) error
	MarkAllRead(ctx context.Context, userId uuid.UUID) error
}

type NotificationHandler struct {
	notificationService notificationService
}

func NewNotificationHandler(notificationService notificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	userId := UserIdFromContext(r.Context())

	limit := utils.GetQueryInt(r, "limit", 10)
	before := utils.GetQueryString(r, "before", "")
	beforeId := utils.GetQueryString(r, "id", "")

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
		ValidationFailedResponse(w, v)
		return
	}

	notifications, err := h.notificationService.List(r.Context(), service.ListNotificationsRequest{
		UserId:          userId,
		BeforeCreatedAt: beforeTime,
		BeforeId:        beforeIdUUID,
		Limit:           limit,
	})
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	if err := utils.WriteJSON(w, http.StatusOK, notifications, nil); err != nil {
		ErrorResponse(w, r, err)
	}
}

func (h *NotificationHandler) CountUnread(w http.ResponseWriter, r *http.Request) {
	userId := UserIdFromContext(r.Context())

	count, err := h.notificationService.CountUnread(r.Context(), userId)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	if err := utils.WriteJSON(w, http.StatusOK, map[string]int{"count": count}, nil); err != nil {
		ErrorResponse(w, r, err)
	}
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userId := UserIdFromContext(r.Context())

	id := chi.URLParam(r, "id")
	notificationId, err := uuid.Parse(id)
	if err != nil {
		BadRequestResponse(w, errors.New("invalid notification id"))
		return
	}

	err = h.notificationService.MarkRead(r.Context(), service.MarkNotificationReadRequest{
		NotificationId: notificationId,
		UserId:         userId,
	})
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	if err := utils.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true}, nil); err != nil {
		ErrorResponse(w, r, err)
	}
}

func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userId := UserIdFromContext(r.Context())

	if err := h.notificationService.MarkAllRead(r.Context(), userId); err != nil {
		ErrorResponse(w, r, err)
		return
	}

	if err := utils.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true}, nil); err != nil {
		ErrorResponse(w, r, err)
	}
}
