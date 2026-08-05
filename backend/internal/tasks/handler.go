package tasks

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/auth"
	platformhttp "github.com/gabrielnakaema/project-chat/internal/platform/http"
	"github.com/gabrielnakaema/project-chat/internal/platform/httperr"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/gabrielnakaema/project-chat/internal/validator"
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
	SearchTasks(ctx context.Context, request SearchTasksRequest) (*utils.CursorPaginated[domain.Task], error)
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
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusCreated, task, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	parsedProjectId, err := platformhttp.ParseOptionalQueryUUID(r, "project_id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}
	if parsedProjectId == nil {
		httperr.BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	projectColumnIDs, err := parseUUIDQueryParam(platformhttp.GetQueryString(r, "project_column_ids", ""))
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	limit, err := platformhttp.ParseLimit(r, 15, 100)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	taskOrder := platformhttp.GetQueryString(r, "task_order", "")

	updatedAt, err := platformhttp.ParseRFC3339Cursor(r, "updated_at")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	userId := auth.UserIdFromContext(r.Context())
	archived := platformhttp.GetQueryString(r, "archived", "") == "true"

	serviceRequest := ListTasksRequest{
		ProjectId:        *parsedProjectId,
		RequestUserId:    userId,
		ProjectColumnIDs: projectColumnIDs,
		Archived:         archived,
		TaskOrder:        taskOrder,
		Limit:            int(limit),
		CursorUpdatedAt:  updatedAt,
	}

	result, err := h.taskService.List(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) GroupByColumn(w http.ResponseWriter, r *http.Request) {
	parsedProjectId, err := platformhttp.ParseOptionalQueryUUID(r, "project_id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}
	if parsedProjectId == nil {
		httperr.BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	limitInt, err := platformhttp.ParseLimit(r, 15, 100)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	taskOrder := platformhttp.GetQueryString(r, "task_order", "")

	updatedAt, err := platformhttp.ParseRFC3339Cursor(r, "updated_at")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	projectColumnIDs, err := parseUUIDQueryParam(platformhttp.GetQueryString(r, "project_column_ids", ""))
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	userId := auth.UserIdFromContext(r.Context())
	archived := platformhttp.GetQueryString(r, "archived", "") == "true"

	serviceRequest := GroupByColumnRequest{
		ProjectId:        *parsedProjectId,
		UserId:           userId,
		ProjectColumnIDs: projectColumnIDs,
		Archived:         archived,
		TaskOrder:        taskOrder,
		Limit:            int(limitInt),
		CursorUpdatedAt:  updatedAt,
	}

	result, err := h.taskService.GroupByColumn(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) CountByColumn(w http.ResponseWriter, r *http.Request) {
	parsedProjectId, err := platformhttp.ParseOptionalQueryUUID(r, "project_id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}
	if parsedProjectId == nil {
		httperr.BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	userId := auth.UserIdFromContext(r.Context())

	projectColumnIDs, err := parseUUIDQueryParam(platformhttp.GetQueryString(r, "project_column_ids", ""))
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	if len(projectColumnIDs) == 0 {
		httperr.BadRequestResponse(w, errors.New("project_column_ids are required"))
		return
	}

	result, err := h.taskService.CountByColumn(r.Context(), *parsedProjectId, projectColumnIDs, userId)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {

	parsedId, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	userId := auth.UserIdFromContext(r.Context())

	task, err := h.taskService.GetById(r.Context(), parsedId, userId)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, task, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	parsedId, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	var request UpdateTaskBody
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
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, task, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) Archive(w http.ResponseWriter, r *http.Request) {
	parsedId, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	userId := auth.UserIdFromContext(r.Context())

	serviceRequest := ArchiveTaskRequest{
		TaskId:        parsedId,
		RequestUserId: userId,
	}

	task, err := h.taskService.Archive(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, task, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) Restore(w http.ResponseWriter, r *http.Request) {
	parsedId, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	var request RestoreTaskBody
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

	userId := auth.UserIdFromContext(r.Context())

	serviceRequest := RestoreTaskRequest{
		TaskId:          parsedId,
		ProjectColumnId: request.ProjectColumnId,
		RequestUserId:   userId,
	}

	task, err := h.taskService.Restore(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, task, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) Move(w http.ResponseWriter, r *http.Request) {
	parsedId, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	var request MoveTaskBody
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
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, task, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
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

	limit, err := platformhttp.ParseLimit(r, 15, 100)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	dueDate, err := platformhttp.ParseRFC3339Cursor(r, "due_date")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	updatedAt, err := platformhttp.ParseRFC3339Cursor(r, "updated_at")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	serviceRequest := ListUserDueTasksRequest{
		UserId:          userId,
		Limit:           int(limit),
		CursorDueDate:   dueDate,
		CursorUpdatedAt: updatedAt,
	}

	result, err := h.taskService.ListUserDueTasks(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) SuggestTaskCodes(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	parsedProjectId, err := platformhttp.ParseOptionalQueryUUID(r, "project_id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}
	if parsedProjectId == nil {
		httperr.BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	prefix := strings.TrimSpace(platformhttp.GetQueryString(r, "prefix", ""))
	if len(prefix) < 2 {
		httperr.BadRequestResponse(w, errors.New("prefix must be at least 2 characters"))
		return
	}

	limit, err := platformhttp.ParseLimit(r, 8, 20)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	result, err := h.taskService.SuggestTaskCodes(r.Context(), SuggestTaskCodesRequest{
		ProjectId: *parsedProjectId,
		UserId:    userId,
		Prefix:    prefix,
		Limit:     int(limit),
	})
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, map[string]any{"data": result}, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *TaskHandler) SearchProjectTasksForDependencies(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	parsedProjectId, err := platformhttp.ParseOptionalQueryUUID(r, "project_id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}
	if parsedProjectId == nil {
		httperr.BadRequestResponse(w, errors.New("project_id is required"))
		return
	}

	searchQuery := strings.TrimSpace(platformhttp.GetQueryString(r, "query", ""))
	if searchQuery == "" {
		httperr.BadRequestResponse(w, errors.New("query is required"))
		return
	}

	limit, err := platformhttp.ParseLimit(r, 20, 100)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	excludeTaskId, err := platformhttp.ParseOptionalQueryUUID(r, "exclude_task_id")
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	result, err := h.taskService.SearchProjectTasksForDependencies(r.Context(), SearchProjectTasksForDependenciesRequest{
		ProjectId:     *parsedProjectId,
		UserId:        userId,
		Query:         searchQuery,
		ExcludeTaskId: excludeTaskId,
		Limit:         int(limit),
	})
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, map[string]any{"data": result}, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func parseSearchTasksRequest(r *http.Request, userId uuid.UUID) (SearchTasksRequest, error) {
	limit, err := platformhttp.ParseLimit(r, 15, 100)
	if err != nil {
		return SearchTasksRequest{}, err
	}

	searchQuery := platformhttp.GetQueryString(r, "query", "")
	if searchQuery == "" {
		return SearchTasksRequest{}, errors.New("query is required")
	}

	dueDate, err := platformhttp.ParseRFC3339Cursor(r, "due_date")
	if err != nil {
		return SearchTasksRequest{}, err
	}

	updatedAt, err := platformhttp.ParseRFC3339Cursor(r, "updated_at")
	if err != nil {
		return SearchTasksRequest{}, err
	}

	taskId, err := platformhttp.ParseOptionalQueryUUID(r, "task_id")
	if err != nil {
		return SearchTasksRequest{}, err
	}

	projectId, err := platformhttp.ParseOptionalQueryUUID(r, "project_id")
	if err != nil {
		return SearchTasksRequest{}, err
	}

	projectColumnIDs, err := parseUUIDQueryParam(platformhttp.GetQueryString(r, "project_column_ids", ""))
	if err != nil {
		return SearchTasksRequest{}, err
	}

	return SearchTasksRequest{
		UserId:           userId,
		ProjectId:        projectId,
		ProjectColumnIDs: projectColumnIDs,
		Limit:            int(limit),
		SearchQuery:      searchQuery,
		IncludeArchived:  platformhttp.GetQueryString(r, "include_archived", "") == "true",
		IncludeDone:      platformhttp.GetQueryString(r, "include_done", "") == "true",
		CursorDueDate:    dueDate,
		CursorUpdatedAt:  updatedAt,
		CursorTaskId:     taskId,
	}, nil
}

func (h *TaskHandler) SearchTasksForUser(w http.ResponseWriter, r *http.Request) {
	serviceRequest, err := parseSearchTasksRequest(r, auth.UserIdFromContext(r.Context()))
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	result, err := h.taskService.SearchTasks(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}
