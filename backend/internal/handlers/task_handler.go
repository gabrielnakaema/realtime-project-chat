package handlers

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

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
	GroupByStatus(ctx context.Context, request service.GroupByStatusRequest) (map[domain.TaskStatus]utils.CursorPaginated[domain.Task], error)
	CountByStatus(ctx context.Context, projectId uuid.UUID, statuses []domain.TaskStatus, requestUserId uuid.UUID) (map[domain.TaskStatus]int, error)
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
		ProjectId:     request.ProjectId,
		Title:         request.Title,
		Description:   request.Description,
		RequestUserId: userId,
		Priority:      request.Priority,
		ResponsibleId: request.ResponsibleId,
		DueDate:       request.DueDate,
		Tags:          request.Tags,
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

	statuses := utils.GetQueryString(r, "statuses", "")
	statusesArray := strings.Split(statuses, ",")
	if statuses == "" {
		statusesArray = []string{}
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

	for _, status := range statusesArray {
		if !slices.Contains(domain.AllowedTaskStatuses, domain.TaskStatus(status)) {
			BadRequestResponse(w, errors.New("invalid status"))
			return
		}

		if status == string(domain.TaskStatusArchived) {
			BadRequestResponse(w, errors.New("archived status is not allowed"))
			return
		}
	}

	taskOrder := utils.GetQueryString(r, "task_order", "0")
	taskOrderInt, err := strconv.Atoi(taskOrder)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	userId := UserIdFromContext(r.Context())

	serviceRequest := service.ListTasksRequest{
		ProjectId:     parsedProjectId,
		RequestUserId: userId,
		Statuses:      statusesArray,
		TaskOrder:     taskOrderInt,
		Limit:         int(limit),
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

func (h *TaskHandler) GroupByStatus(w http.ResponseWriter, r *http.Request) {
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

	limit := utils.GetQueryString(r, "limit", "15")
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	if limitInt < 1 {
		BadRequestResponse(w, errors.New("limit must be greater than 0"))
		return
	}

	if limitInt > 100 {
		BadRequestResponse(w, errors.New("limit must be less than 100"))
		return
	}

	taskOrder := utils.GetQueryString(r, "task_order", "0")
	taskOrderInt, err := strconv.Atoi(taskOrder)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}
	if taskOrderInt < 0 {
		BadRequestResponse(w, errors.New("task_order must be greater than 0"))
		return
	}

	statuses := utils.GetQueryString(r, "statuses", "")
	statusesArray := []domain.TaskStatus{}
	for _, status := range strings.Split(statuses, ",") {
		if !slices.Contains(domain.AllowedTaskStatuses, domain.TaskStatus(status)) {
			BadRequestResponse(w, errors.New("invalid status"))
			return
		}

		if status == string(domain.TaskStatusArchived) {
			BadRequestResponse(w, errors.New("archived status is not allowed"))
			return
		}
		statusesArray = append(statusesArray, domain.TaskStatus(status))
	}

	userId := UserIdFromContext(r.Context())

	serviceRequest := service.GroupByStatusRequest{
		ProjectId: parsedProjectId,
		UserId:    userId,
		Statuses:  statusesArray,
		TaskOrder: taskOrderInt,
		Limit:     limitInt,
	}

	result, err := h.taskService.GroupByStatus(r.Context(), serviceRequest)
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

func (h *TaskHandler) CountByStatus(w http.ResponseWriter, r *http.Request) {
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

	statuses := utils.GetQueryString(r, "statuses", "")
	statusesArray := []domain.TaskStatus{}
	for _, status := range strings.Split(statuses, ",") {
		if !slices.Contains(domain.AllowedTaskStatuses, domain.TaskStatus(status)) {
			BadRequestResponse(w, errors.New("invalid status"))
			return
		}
		statusesArray = append(statusesArray, domain.TaskStatus(status))
	}

	if len(statusesArray) == 0 {
		BadRequestResponse(w, errors.New("statuses are required"))
		return
	}

	result, err := h.taskService.CountByStatus(r.Context(), parsedProjectId, statusesArray, userId)
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
		TaskId:        parsedId,
		Title:         request.Title,
		Description:   request.Description,
		Status:        domain.TaskStatus(request.Status),
		RequestUserId: userId,
		Priority:      domain.TaskPriority(request.Priority),
		ResponsibleId: request.ResponsibleId,
		DueDate:       request.DueDate,
		Tags:          request.Tags,
		DoneAt:        request.DoneAt,
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
		TaskId:        parsedId,
		RequestUserId: userId,
		AfterTaskId:   request.AfterTaskId,
		ProjectId:     request.ProjectId,
		Status:        domain.TaskStatus(request.Status),
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
