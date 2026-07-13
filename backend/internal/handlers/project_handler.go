package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/auth"
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
	UpdateColumn(ctx context.Context, request service.UpdateProjectColumnRequest) (*domain.ProjectColumn, error)
	CreateMember(ctx context.Context, request service.CreateMemberRequest) (*domain.ProjectMember, error)
	ListActivitiesByProject(ctx context.Context, request service.ListActivitiesByProjectRequest) (*utils.CursorPaginated[domain.ProjectActivity], error)
	ListUsersProjectActivities(ctx context.Context, request service.ListUsersProjectActivitiesRequest) (*utils.CursorPaginated[domain.ProjectActivity], error)
	ListMembersByProjectId(ctx context.Context, projectId uuid.UUID, requestUserId uuid.UUID) ([]domain.ProjectMember, error)
	RemoveMember(ctx context.Context, request service.RemoveMemberRequest) error
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

	userId := auth.UserIdFromContext(r.Context())

	serviceRequest := service.CreateProjectRequest{
		Name:             request.Name,
		Description:      request.Description,
		RepositoryURL:    request.RepositoryURL,
		RepositoryOwner:  request.RepositoryOwner,
		RepositoryName:   request.RepositoryName,
		DefaultBranch:    request.DefaultBranch,
		BranchNamePrefix: request.BranchNamePrefix,
		UserId:           userId,
		Columns:          mapProjectColumnsRequest(request.Columns),
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
	userId := auth.UserIdFromContext(r.Context())

	memberRole := utils.GetQueryString(r, "member_role", "")

	v := validator.New()
	if memberRole != "" {
		v.Check("member_role", "member_role is invalid", memberRole == string(domain.ProjectMemberRoleMember) || memberRole == string(domain.ProjectMemberRoleCreator))
	}

	searchQuery := utils.GetQueryString(r, "query", "")
	if searchQuery != "" {
		v.Check("query", "query must be less than 100 characters", len(searchQuery) <= 100)
	}

	if !v.Valid() {
		ValidationFailedResponse(w, v)
		return
	}

	serviceRequest := service.ListProjectsByUserIdRequest{
		UserId:             userId,
		MemberRole:         domain.ProjectMemberRole(memberRole),
		ShouldFilterByRole: memberRole != "",
		SearchQuery:        searchQuery,
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
	userId := auth.UserIdFromContext(r.Context())

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
	userId := auth.UserIdFromContext(r.Context())

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
		Id:               parsed,
		Name:             request.Name,
		Description:      request.Description,
		RepositoryURL:    request.RepositoryURL,
		RepositoryOwner:  request.RepositoryOwner,
		RepositoryName:   request.RepositoryName,
		DefaultBranch:    request.DefaultBranch,
		BranchNamePrefix: request.BranchNamePrefix,
		UserId:           userId,
		Columns:          mapProjectColumnsRequest(request.Columns),
		DeletedColumns:   mapDeletedProjectColumnsRequest(request.DeletedColumns),
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

func (h *ProjectHandler) UpdateColumn(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	projectID := chi.URLParam(r, "id")
	parsedProjectID, err := uuid.Parse(projectID)
	if err != nil {
		BadRequestResponse(w, errors.New("invalid project id"))
		return
	}

	columnID := chi.URLParam(r, "column_id")
	parsedColumnID, err := uuid.Parse(columnID)
	if err != nil {
		BadRequestResponse(w, errors.New("invalid project column id"))
		return
	}

	var request UpdateProjectColumnRequest
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

	column, err := h.projectService.UpdateColumn(r.Context(), service.UpdateProjectColumnRequest{
		ProjectId:    parsedProjectID,
		ColumnId:     parsedColumnID,
		UserId:       userId,
		Name:         request.Name,
		Description:  request.Description,
		Color:        request.Color,
		IsDoneColumn: request.IsDoneColumn,
	})
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, column, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func mapProjectColumnsRequest(columns []ProjectColumnRequest) []domain.ProjectColumn {
	mapped := make([]domain.ProjectColumn, len(columns))
	for i, column := range columns {
		if column.Id != nil {
			mapped[i].Id = *column.Id
		}
		mapped[i].Name = column.Name
		mapped[i].Description = column.Description
		mapped[i].Color = column.Color
		mapped[i].Position = i
		mapped[i].IsDoneColumn = column.IsDoneColumn
	}

	return mapped
}

func mapDeletedProjectColumnsRequest(columns []DeletedProjectColumnRequest) []service.DeletedProjectColumnRequest {
	mapped := make([]service.DeletedProjectColumnRequest, len(columns))
	for i, column := range columns {
		mapped[i] = service.DeletedProjectColumnRequest{
			Id:                  column.Id,
			MoveTasksToColumnId: column.MoveTasksToColumnId,
		}
	}

	return mapped
}

func (h *ProjectHandler) CreateMember(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

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

func (h *ProjectHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	id := chi.URLParam(r, "id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		BadRequestResponse(w, errors.New("invalid project id"))
		return
	}

	memberId := chi.URLParam(r, "member_id")
	parsedMemberId, err := uuid.Parse(memberId)
	if err != nil {
		BadRequestResponse(w, errors.New("invalid member id"))
		return
	}

	serviceRequest := service.RemoveMemberRequest{
		ProjectId:     parsed,
		UserId:        parsedMemberId,
		RequestUserId: userId,
	}

	err = h.projectService.RemoveMember(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusNoContent, nil, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) ListActivitiesByProject(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	id := chi.URLParam(r, "id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		BadRequestResponse(w, errors.New("invalid project id"))
		return
	}

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

	serviceRequest := service.ListActivitiesByProjectRequest{
		ProjectId:       parsed,
		UserId:          userId,
		BeforeCreatedAt: beforeTime,
		BeforeId:        beforeIdUUID,
		Limit:           int32(limit),
	}

	activities, err := h.projectService.ListActivitiesByProject(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, activities, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) ListUsersProjectActivities(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

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

	serviceRequest := service.ListUsersProjectActivitiesRequest{
		UserId:          userId,
		BeforeCreatedAt: beforeTime,
		BeforeId:        beforeIdUUID,
		Limit:           int32(limit),
	}

	activities, err := h.projectService.ListUsersProjectActivities(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, activities, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) ListMembersByProjectId(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	id := chi.URLParam(r, "id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		BadRequestResponse(w, errors.New("invalid project id"))
		return
	}

	members, err := h.projectService.ListMembersByProjectId(r.Context(), parsed, userId)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, members, nil)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}
}
