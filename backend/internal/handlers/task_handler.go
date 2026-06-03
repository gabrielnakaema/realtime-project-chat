package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type taskService interface {
	Create(ctx context.Context, request service.CreateTaskRequest) (*domain.Task, error)
	List(ctx context.Context, request service.ListTasksRequest) (*utils.CursorPaginated[domain.Task], error)
	GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Task, error)
	Update(ctx context.Context, request service.UpdateTaskRequest) (*domain.Task, error)
	Move(ctx context.Context, request service.MoveTaskRequest) (*domain.Task, error)
	GroupByColumn(ctx context.Context, request service.GroupByColumnRequest) (map[string]utils.CursorPaginated[domain.Task], error)
	CountByColumn(ctx context.Context, projectId uuid.UUID, projectColumnIDs []uuid.UUID, requestUserId uuid.UUID) (map[string]int, error)
	Archive(ctx context.Context, request service.ArchiveTaskRequest) (*domain.Task, error)
	Restore(ctx context.Context, request service.RestoreTaskRequest) (*domain.Task, error)
	ListUserDueTasks(ctx context.Context, request service.ListUserDueTasksRequest) (*utils.CursorPaginated[domain.Task], error)
	SearchTasksForUser(ctx context.Context, request service.SearchTasksForUserRequest) (*utils.CursorPaginated[domain.Task], error)
}

type TaskHandler struct {
	taskService taskService
}

func NewTaskHandler(taskService taskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {

	var request CreateTaskRequest
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

	serviceRequest := service.CreateTaskRequest{
		ProjectId:       request.ProjectId,
		ProjectColumnId: request.ProjectColumnId,
		Title:           request.Title,
		Description:     request.Description,
		Code:            request.Code,
		RequestUserId:   userId,
		Priority:        request.Priority,
		ResponsibleId:   request.ResponsibleId,
		DueDate:         request.DueDate,
		Tags:            request.Tags,
	}

	task, err := h.taskService.Create(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, task, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	projectId := utils.GetQueryString(r, "project_id", "")
	if projectId == "" {
		BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	parsedProjectId, err := uuid.Parse(projectId)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	projectColumnIDs, err := parseUUIDQueryParam(utils.GetQueryString(r, "project_column_ids", ""))
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	limit := utils.GetQueryInt(r, "limit", 15)
	if limit <= 0 {
		BadRequestResponse(w, errors.New("limit must be greater than 0"))
		return
	}

	if limit > 100 {
		BadRequestResponse(w, errors.New("limit must be less than 100"))
		return
	}

	taskOrder := utils.GetQueryString(r, "task_order", "")

	cursorUpdatedAt := utils.GetQueryString(r, "updated_at", "")
	var updatedAt *time.Time
	if cursorUpdatedAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, cursorUpdatedAt)
		if err != nil {
			BadRequestResponse(w, err)
			return
		}
		updatedAt = &parsedTime
	}

	userId := UserIdFromContext(r.Context())
	archived := utils.GetQueryString(r, "archived", "") == "true"

	serviceRequest := service.ListTasksRequest{
		ProjectId:        parsedProjectId,
		RequestUserId:    userId,
		ProjectColumnIDs: projectColumnIDs,
		Archived:         archived,
		TaskOrder:        taskOrder,
		Limit:            int(limit),
		CursorUpdatedAt:  updatedAt,
	}

	result, err := h.taskService.List(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) GroupByColumn(w http.ResponseWriter, r *http.Request) {
	projectId := utils.GetQueryString(r, "project_id", "")
	if projectId == "" {
		BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	parsedProjectId, err := uuid.Parse(projectId)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	limitInt := utils.GetQueryInt(r, "limit", 15)

	if limitInt < 1 {
		BadRequestResponse(w, errors.New("limit must be greater than 0"))
		return
	}

	if limitInt > 100 {
		BadRequestResponse(w, errors.New("limit must be less than 100"))
		return
	}

	taskOrder := utils.GetQueryString(r, "task_order", "")

	cursorUpdatedAt := utils.GetQueryString(r, "updated_at", "")
	var updatedAt *time.Time
	if cursorUpdatedAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, cursorUpdatedAt)
		if err != nil {
			BadRequestResponse(w, err)
			return
		}
		updatedAt = &parsedTime
	}

	projectColumnIDs, err := parseUUIDQueryParam(utils.GetQueryString(r, "project_column_ids", ""))
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	userId := UserIdFromContext(r.Context())
	archived := utils.GetQueryString(r, "archived", "") == "true"

	serviceRequest := service.GroupByColumnRequest{
		ProjectId:        parsedProjectId,
		UserId:           userId,
		ProjectColumnIDs: projectColumnIDs,
		Archived:         archived,
		TaskOrder:        taskOrder,
		Limit:            int(limitInt),
		CursorUpdatedAt:  updatedAt,
	}

	result, err := h.taskService.GroupByColumn(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) CountByColumn(w http.ResponseWriter, r *http.Request) {
	projectId := utils.GetQueryString(r, "project_id", "")
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

	projectColumnIDs, err := parseUUIDQueryParam(utils.GetQueryString(r, "project_column_ids", ""))
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	if len(projectColumnIDs) == 0 {
		BadRequestResponse(w, errors.New("project_column_ids are required"))
		return
	}

	result, err := h.taskService.CountByColumn(r.Context(), parsedProjectId, projectColumnIDs, userId)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	userId := UserIdFromContext(r.Context())

	task, err := h.taskService.GetById(r.Context(), parsedId, userId)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, task, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	var request UpdateTaskRequest
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

	userId := UserIdFromContext(r.Context())

	serviceRequest := service.UpdateTaskRequest{
		TaskId:          parsedId,
		Title:           request.Title,
		Description:     request.Description,
		Code:            request.Code,
		ProjectColumnId: request.ProjectColumnId,
		RequestUserId:   userId,
		Priority:        domain.TaskPriority(request.Priority),
		ResponsibleId:   request.ResponsibleId,
		DueDate:         request.DueDate,
		Tags:            request.Tags,
	}

	task, err := h.taskService.Update(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, task, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) Archive(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	userId := UserIdFromContext(r.Context())

	serviceRequest := service.ArchiveTaskRequest{
		TaskId:        parsedId,
		RequestUserId: userId,
	}

	task, err := h.taskService.Archive(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, task, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) Restore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	var request RestoreTaskRequest
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

	userId := UserIdFromContext(r.Context())

	serviceRequest := service.RestoreTaskRequest{
		TaskId:          parsedId,
		ProjectColumnId: request.ProjectColumnId,
		RequestUserId:   userId,
	}

	task, err := h.taskService.Restore(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, task, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) Move(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	var request MoveTaskRequest
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

	userId := UserIdFromContext(r.Context())

	serviceRequest := service.MoveTaskRequest{
		TaskId:          parsedId,
		RequestUserId:   userId,
		AfterTaskId:     request.AfterTaskId,
		ProjectId:       request.ProjectId,
		ProjectColumnId: request.ProjectColumnId,
	}

	task, err := h.taskService.Move(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, task, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func parseUUIDQueryParam(value string) ([]uuid.UUID, error) {
	if value == "" {
		return []uuid.UUID{}, nil
	}

	parts := strings.Split(value, ",")
	ids := make([]uuid.UUID, 0, len(parts))
	for _, part := range parts {
		parsed, err := uuid.Parse(part)
		if err != nil {
			return nil, errors.New("invalid project_column_id")
		}
		ids = append(ids, parsed)
	}

	return ids, nil
}

func (h *TaskHandler) ListUserDueTasks(w http.ResponseWriter, r *http.Request) {
	userId := UserIdFromContext(r.Context())

	limit := utils.GetQueryInt(r, "limit", 15)
	if limit <= 0 {
		BadRequestResponse(w, errors.New("limit must be greater than 0"))
		return
	}

	if limit > 100 {
		BadRequestResponse(w, errors.New("limit must be less than 100"))
		return
	}

	cursorDueDate := utils.GetQueryString(r, "due_date", "")
	var dueDate *time.Time
	if cursorDueDate != "" {
		parsedTime, err := time.Parse(time.RFC3339, cursorDueDate)
		if err != nil {
			BadRequestResponse(w, err)
			return
		}
		dueDate = &parsedTime
	}

	cursorUpdatedAt := utils.GetQueryString(r, "updated_at", "")
	var updatedAt *time.Time
	if cursorUpdatedAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, cursorUpdatedAt)
		if err != nil {
			BadRequestResponse(w, err)
			return
		}
		updatedAt = &parsedTime
	}

	serviceRequest := service.ListUserDueTasksRequest{
		UserId:          userId,
		Limit:           int(limit),
		CursorDueDate:   dueDate,
		CursorUpdatedAt: updatedAt,
	}

	result, err := h.taskService.ListUserDueTasks(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) SearchTasksForUser(w http.ResponseWriter, r *http.Request) {
	userId := UserIdFromContext(r.Context())

	limit := utils.GetQueryInt(r, "limit", 15)
	if limit <= 0 {
		BadRequestResponse(w, errors.New("limit must be greater than 0"))
		return
	}

	if limit > 100 {
		BadRequestResponse(w, errors.New("limit must be less than 100"))
		return
	}

	searchQuery := utils.GetQueryString(r, "query", "")
	if searchQuery == "" {
		BadRequestResponse(w, errors.New("query is required"))
		return
	}

	cursorDueDate := utils.GetQueryString(r, "due_date", "")
	var dueDate *time.Time
	if cursorDueDate != "" {
		parsedTime, err := time.Parse(time.RFC3339, cursorDueDate)
		if err != nil {
			BadRequestResponse(w, err)
			return
		}
		dueDate = &parsedTime
	}

	cursorUpdatedAt := utils.GetQueryString(r, "updated_at", "")
	var updatedAt *time.Time
	if cursorUpdatedAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, cursorUpdatedAt)
		if err != nil {
			BadRequestResponse(w, err)
			return
		}
		updatedAt = &parsedTime
	}

	serviceRequest := service.SearchTasksForUserRequest{
		UserId:          userId,
		Limit:           int(limit),
		SearchQuery:     searchQuery,
		CursorDueDate:   dueDate,
		CursorUpdatedAt: updatedAt,
	}

	result, err := h.taskService.SearchTasksForUser(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}
