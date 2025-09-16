package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type taskService interface {
	Create(ctx context.Context, request service.CreateTaskRequest) (*domain.Task, error)
	List(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) ([]domain.Task, error)
	GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Task, error)
	Update(ctx context.Context, request service.UpdateTaskRequest) (*domain.Task, error)
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

	userId := UserIdFromContext(r.Context())

	tasks, err := h.taskService.List(r.Context(), parsedProjectId, userId)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	type response struct {
		Data []domain.Task `json:"data"`
	}

	err = utils.WriteJSON(w, http.StatusOK, response{Data: tasks}, nil)
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
