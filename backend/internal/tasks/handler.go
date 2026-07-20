package tasks

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/auth"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/handlers"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type taskService interface {
	Create(ctx context.Context, request CreateTaskRequest) (*domain.Task, error)
	List(ctx context.Context, request ListTasksRequest) (*utils.CursorPaginated[domain.Task], error)
	GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Task, error)
	Update(ctx context.Context, request UpdateTaskRequest) (*domain.Task, error)
	Move(ctx context.Context, request MoveTaskRequest) (*domain.Task, error)
	GroupByColumn(ctx context.Context, request GroupByColumnRequest) (map[string]utils.CursorPaginated[domain.Task], error)
	CountByColumn(ctx context.Context, projectId uuid.UUID, projectColumnIDs []uuid.UUID, requestUserId uuid.UUID) (map[string]int, error)
	Archive(ctx context.Context, request ArchiveTaskRequest) (*domain.Task, error)
	Restore(ctx context.Context, request RestoreTaskRequest) (*domain.Task, error)
	ListUserDueTasks(ctx context.Context, request ListUserDueTasksRequest) (*utils.CursorPaginated[domain.Task], error)
	SearchTasksForUser(ctx context.Context, request SearchTasksForUserRequest) (*utils.CursorPaginated[domain.Task], error)
	SuggestTaskCodes(ctx context.Context, request SuggestTaskCodesRequest) ([]domain.TaskCodeSuggestion, error)
	SearchProjectTasksForDependencies(ctx context.Context, request SearchProjectTasksForDependenciesRequest) ([]domain.TaskDependencyRef, error)
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

	var request CreateTaskBody
	err := utils.ReadJSON(w, r, &request)
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

	userId := auth.UserIdFromContext(r.Context())

	serviceRequest := CreateTaskRequest{
		ProjectId:        request.ProjectId,
		ProjectColumnId:  request.ProjectColumnId,
		Title:            request.Title,
		Description:      request.Description,
		Code:             request.Code,
		RequestUserId:    userId,
		Priority:         request.Priority,
		ResponsibleId:    request.ResponsibleId,
		DueDate:          request.DueDate,
		Tags:             request.Tags,
		DependsOnTaskIds: request.DependsOnTaskIds,
	}

	task, err := h.taskService.Create(r.Context(), serviceRequest)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, task, nil)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	projectId := utils.GetQueryString(r, "project_id", "")
	if projectId == "" {
		handlers.BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	parsedProjectId, err := uuid.Parse(projectId)
	if err != nil {
		handlers.BadRequestResponse(w, err)
		return
	}

	projectColumnIDs, err := parseUUIDQueryParam(utils.GetQueryString(r, "project_column_ids", ""))
	if err != nil {
		handlers.BadRequestResponse(w, err)
		return
	}

	limit := utils.GetQueryInt(r, "limit", 15)
	if limit <= 0 {
		handlers.BadRequestResponse(w, errors.New("limit must be greater than 0"))
		return
	}

	if limit > 100 {
		handlers.BadRequestResponse(w, errors.New("limit must be less than 100"))
		return
	}

	taskOrder := utils.GetQueryString(r, "task_order", "")

	cursorUpdatedAt := utils.GetQueryString(r, "updated_at", "")
	var updatedAt *time.Time
	if cursorUpdatedAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, cursorUpdatedAt)
		if err != nil {
			handlers.BadRequestResponse(w, err)
			return
		}
		updatedAt = &parsedTime
	}

	userId := auth.UserIdFromContext(r.Context())
	archived := utils.GetQueryString(r, "archived", "") == "true"

	serviceRequest := ListTasksRequest{
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
		handlers.ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) GroupByColumn(w http.ResponseWriter, r *http.Request) {
	projectId := utils.GetQueryString(r, "project_id", "")
	if projectId == "" {
		handlers.BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	parsedProjectId, err := uuid.Parse(projectId)
	if err != nil {
		handlers.BadRequestResponse(w, err)
		return
	}

	limitInt := utils.GetQueryInt(r, "limit", 15)

	if limitInt < 1 {
		handlers.BadRequestResponse(w, errors.New("limit must be greater than 0"))
		return
	}

	if limitInt > 100 {
		handlers.BadRequestResponse(w, errors.New("limit must be less than 100"))
		return
	}

	taskOrder := utils.GetQueryString(r, "task_order", "")

	cursorUpdatedAt := utils.GetQueryString(r, "updated_at", "")
	var updatedAt *time.Time
	if cursorUpdatedAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, cursorUpdatedAt)
		if err != nil {
			handlers.BadRequestResponse(w, err)
			return
		}
		updatedAt = &parsedTime
	}

	projectColumnIDs, err := parseUUIDQueryParam(utils.GetQueryString(r, "project_column_ids", ""))
	if err != nil {
		handlers.BadRequestResponse(w, err)
		return
	}

	userId := auth.UserIdFromContext(r.Context())
	archived := utils.GetQueryString(r, "archived", "") == "true"

	serviceRequest := GroupByColumnRequest{
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
		handlers.ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) CountByColumn(w http.ResponseWriter, r *http.Request) {
	projectId := utils.GetQueryString(r, "project_id", "")
	if projectId == "" {
		handlers.BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	parsedProjectId, err := uuid.Parse(projectId)
	if err != nil {
		handlers.BadRequestResponse(w, err)
		return
	}

	userId := auth.UserIdFromContext(r.Context())

	projectColumnIDs, err := parseUUIDQueryParam(utils.GetQueryString(r, "project_column_ids", ""))
	if err != nil {
		handlers.BadRequestResponse(w, err)
		return
	}

	if len(projectColumnIDs) == 0 {
		handlers.BadRequestResponse(w, errors.New("project_column_ids are required"))
		return
	}

	result, err := h.taskService.CountByColumn(r.Context(), parsedProjectId, projectColumnIDs, userId)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		handlers.BadRequestResponse(w, err)
		return
	}

	userId := auth.UserIdFromContext(r.Context())

	task, err := h.taskService.GetById(r.Context(), parsedId, userId)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, task, nil)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		handlers.BadRequestResponse(w, err)
		return
	}

	var request UpdateTaskBody
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

	userId := auth.UserIdFromContext(r.Context())

	serviceRequest := UpdateTaskRequest{
		TaskId:           parsedId,
		Title:            request.Title,
		Description:      request.Description,
		Code:             &request.Code,
		ProjectColumnId:  request.ProjectColumnId,
		RequestUserId:    userId,
		Priority:         domain.TaskPriority(request.Priority),
		ResponsibleId:    request.ResponsibleId,
		DueDate:          request.DueDate,
		Tags:             request.Tags,
		DependsOnTaskIds: request.DependsOnTaskIds,
	}

	task, err := h.taskService.Update(r.Context(), serviceRequest)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, task, nil)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) Archive(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		handlers.BadRequestResponse(w, err)
		return
	}

	userId := auth.UserIdFromContext(r.Context())

	serviceRequest := ArchiveTaskRequest{
		TaskId:        parsedId,
		RequestUserId: userId,
	}

	task, err := h.taskService.Archive(r.Context(), serviceRequest)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, task, nil)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) Restore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		handlers.BadRequestResponse(w, err)
		return
	}

	var request RestoreTaskBody
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

	userId := auth.UserIdFromContext(r.Context())

	serviceRequest := RestoreTaskRequest{
		TaskId:          parsedId,
		ProjectColumnId: request.ProjectColumnId,
		RequestUserId:   userId,
	}

	task, err := h.taskService.Restore(r.Context(), serviceRequest)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, task, nil)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) Move(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		handlers.BadRequestResponse(w, err)
		return
	}

	var request MoveTaskBody
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

	userId := auth.UserIdFromContext(r.Context())

	serviceRequest := MoveTaskRequest{
		TaskId:          parsedId,
		RequestUserId:   userId,
		AfterTaskId:     request.AfterTaskId,
		ProjectId:       request.ProjectId,
		ProjectColumnId: request.ProjectColumnId,
	}

	task, err := h.taskService.Move(r.Context(), serviceRequest)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, task, nil)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
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
	userId := auth.UserIdFromContext(r.Context())

	limit := utils.GetQueryInt(r, "limit", 15)
	if limit <= 0 {
		handlers.BadRequestResponse(w, errors.New("limit must be greater than 0"))
		return
	}

	if limit > 100 {
		handlers.BadRequestResponse(w, errors.New("limit must be less than 100"))
		return
	}

	cursorDueDate := utils.GetQueryString(r, "due_date", "")
	var dueDate *time.Time
	if cursorDueDate != "" {
		parsedTime, err := time.Parse(time.RFC3339, cursorDueDate)
		if err != nil {
			handlers.BadRequestResponse(w, err)
			return
		}
		dueDate = &parsedTime
	}

	cursorUpdatedAt := utils.GetQueryString(r, "updated_at", "")
	var updatedAt *time.Time
	if cursorUpdatedAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, cursorUpdatedAt)
		if err != nil {
			handlers.BadRequestResponse(w, err)
			return
		}
		updatedAt = &parsedTime
	}

	serviceRequest := ListUserDueTasksRequest{
		UserId:          userId,
		Limit:           int(limit),
		CursorDueDate:   dueDate,
		CursorUpdatedAt: updatedAt,
	}

	result, err := h.taskService.ListUserDueTasks(r.Context(), serviceRequest)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) SuggestTaskCodes(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	projectId := utils.GetQueryString(r, "project_id", "")
	if projectId == "" {
		handlers.BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	parsedProjectId, err := uuid.Parse(projectId)
	if err != nil {
		handlers.BadRequestResponse(w, err)
		return
	}

	prefix := strings.TrimSpace(utils.GetQueryString(r, "prefix", ""))
	if len(prefix) < 2 {
		handlers.BadRequestResponse(w, errors.New("prefix must be at least 2 characters"))
		return
	}

	limit := utils.GetQueryInt(r, "limit", 8)
	if limit <= 0 {
		handlers.BadRequestResponse(w, errors.New("limit must be greater than 0"))
		return
	}

	if limit > 20 {
		handlers.BadRequestResponse(w, errors.New("limit must be at most 20"))
		return
	}

	result, err := h.taskService.SuggestTaskCodes(r.Context(), SuggestTaskCodesRequest{
		ProjectId: parsedProjectId,
		UserId:    userId,
		Prefix:    prefix,
		Limit:     int(limit),
	})
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, map[string]any{"data": result}, nil)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) SearchProjectTasksForDependencies(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	projectId := utils.GetQueryString(r, "project_id", "")
	if projectId == "" {
		handlers.BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	parsedProjectId, err := uuid.Parse(projectId)
	if err != nil {
		handlers.BadRequestResponse(w, err)
		return
	}

	searchQuery := strings.TrimSpace(utils.GetQueryString(r, "query", ""))
	if searchQuery == "" {
		handlers.BadRequestResponse(w, errors.New("query is required"))
		return
	}

	limit := utils.GetQueryInt(r, "limit", 20)
	if limit <= 0 {
		handlers.BadRequestResponse(w, errors.New("limit must be greater than 0"))
		return
	}

	if limit > 100 {
		handlers.BadRequestResponse(w, errors.New("limit must be less than 100"))
		return
	}

	var excludeTaskId *uuid.UUID
	excludeTaskIdParam := utils.GetQueryString(r, "exclude_task_id", "")
	if excludeTaskIdParam != "" {
		parsedExcludeTaskId, parseErr := uuid.Parse(excludeTaskIdParam)
		if parseErr != nil {
			handlers.BadRequestResponse(w, parseErr)
			return
		}
		excludeTaskId = &parsedExcludeTaskId
	}

	result, err := h.taskService.SearchProjectTasksForDependencies(r.Context(), SearchProjectTasksForDependenciesRequest{
		ProjectId:     parsedProjectId,
		UserId:        userId,
		Query:         searchQuery,
		ExcludeTaskId: excludeTaskId,
		Limit:         int(limit),
	})
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, map[string]any{"data": result}, nil)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) SearchTasksForUser(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	limit := utils.GetQueryInt(r, "limit", 15)
	if limit <= 0 {
		handlers.BadRequestResponse(w, errors.New("limit must be greater than 0"))
		return
	}

	if limit > 100 {
		handlers.BadRequestResponse(w, errors.New("limit must be less than 100"))
		return
	}

	searchQuery := utils.GetQueryString(r, "query", "")
	if searchQuery == "" {
		handlers.BadRequestResponse(w, errors.New("query is required"))
		return
	}

	cursorDueDate := utils.GetQueryString(r, "due_date", "")
	var dueDate *time.Time
	if cursorDueDate != "" {
		parsedTime, err := time.Parse(time.RFC3339, cursorDueDate)
		if err != nil {
			handlers.BadRequestResponse(w, err)
			return
		}
		dueDate = &parsedTime
	}

	cursorUpdatedAt := utils.GetQueryString(r, "updated_at", "")
	var updatedAt *time.Time
	if cursorUpdatedAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, cursorUpdatedAt)
		if err != nil {
			handlers.BadRequestResponse(w, err)
			return
		}
		updatedAt = &parsedTime
	}

	serviceRequest := SearchTasksForUserRequest{
		UserId:          userId,
		Limit:           int(limit),
		SearchQuery:     searchQuery,
		CursorDueDate:   dueDate,
		CursorUpdatedAt: updatedAt,
	}

	result, err := h.taskService.SearchTasksForUser(r.Context(), serviceRequest)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		handlers.ErrorResponse(w, r, err)
		return
	}
}
