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

type projectService interface {
	Create(ctx context.Context, request service.CreateProjectRequest) (*domain.Project, error)
	GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Project, error)
	ListByUserId(ctx context.Context, request service.ListProjectsByUserIdRequest) ([]domain.Project, error)
	Update(ctx context.Context, request service.UpdateProjectRequest) (*domain.Project, error)
	CreateMember(ctx context.Context, request service.CreateMemberRequest) (*domain.ProjectMember, error)
}

type ProjectHandler struct {
	projectService projectService
}

func NewProjectHandler(projectService projectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request ProjectRequest
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

	serviceRequest := service.CreateProjectRequest{
		Name:        request.Name,
		Description: request.Description,
		UserId:      userId,
	}

	project, err := h.projectService.Create(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, project, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userId := UserIdFromContext(r.Context())

	memberRole := utils.GetQueryString(r, "member_role", "")

	v := validator.New()
	if memberRole != "" {
		v.Check("member_role", "member_role is required", memberRole == string(domain.ProjectMemberRoleMember) || memberRole == string(domain.ProjectMemberRoleCreator))
	}

	serviceRequest := service.ListProjectsByUserIdRequest{
		UserId:             userId,
		MemberRole:         domain.ProjectMemberRole(memberRole),
		ShouldFilterByRole: memberRole != "",
	}

	projects, err := h.projectService.ListByUserId(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, projects, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	userId := UserIdFromContext(r.Context())

	id := chi.URLParam(r, "id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		BadRequestResponse(w, errors.New("invalid project id"))
		return
	}

	project, err := h.projectService.GetById(r.Context(), parsed, userId)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, project, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	userId := UserIdFromContext(r.Context())

	id := chi.URLParam(r, "id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		BadRequestResponse(w, errors.New("invalid project id"))
		return
	}

	var request ProjectRequest
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

	serviceRequest := service.UpdateProjectRequest{
		Id:          parsed,
		Name:        request.Name,
		Description: request.Description,
		UserId:      userId,
	}

	project, err := h.projectService.Update(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, project, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) CreateMember(w http.ResponseWriter, r *http.Request) {
	userId := UserIdFromContext(r.Context())

	id := chi.URLParam(r, "id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		BadRequestResponse(w, errors.New("invalid project id"))
		return
	}

	var request CreateMemberRequest
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

	serviceRequest := service.CreateMemberRequest{
		Email:         request.Email,
		ProjectId:     parsed,
		RequestUserId: userId,
	}

	member, err := h.projectService.CreateMember(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusCreated, member, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}
