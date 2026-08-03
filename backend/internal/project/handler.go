package project

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
	"github.com/google/uuid"
)

type projectService interface {
	Create(ctx context.Context, request CreateProjectRequest) (*domain.Project, error)
	GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Project, error)
	ListByUserId(ctx context.Context, request ListProjectsByUserIdRequest) ([]domain.Project, error)
	Update(ctx context.Context, request UpdateProjectRequest) (*domain.Project, error)
	UpdateColumn(ctx context.Context, request UpdateProjectColumnRequest) (*domain.ProjectColumn, error)
	CreateMember(ctx context.Context, request CreateMemberRequest) (*domain.ProjectMember, error)
	ListActivitiesByProject(ctx context.Context, request ListActivitiesByProjectRequest) (*utils.CursorPaginated[domain.ProjectActivity], error)
	ListUsersProjectActivities(ctx context.Context, request ListUsersProjectActivitiesRequest) (*utils.CursorPaginated[domain.ProjectActivity], error)
	ListMembersByProjectId(ctx context.Context, projectId uuid.UUID, requestUserId uuid.UUID) ([]domain.ProjectMember, error)
	RemoveMember(ctx context.Context, request RemoveMemberRequest) error
	Delete(ctx context.Context, request DeleteProjectRequest) error
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
	var request ProjectBody
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

	serviceRequest := CreateProjectRequest{
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
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusCreated, project, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	memberRole := platformhttp.GetQueryString(r, "member_role", "")

	v := validator.New()
	if memberRole != "" {
		v.Check("member_role", "member_role is invalid", memberRole == string(domain.ProjectMemberRoleMember) || memberRole == string(domain.ProjectMemberRoleCreator))
	}

	searchQuery := platformhttp.GetQueryString(r, "query", "")
	if searchQuery != "" {
		v.Check("query", "query must be less than 100 characters", len(searchQuery) <= 100)
	}

	if !v.Valid() {
		httperr.ValidationFailedResponse(w, v)
		return
	}

	serviceRequest := ListProjectsByUserIdRequest{
		UserId:             userId,
		MemberRole:         domain.ProjectMemberRole(memberRole),
		ShouldFilterByRole: memberRole != "",
		SearchQuery:        searchQuery,
	}

	projects, err := h.projectService.ListByUserId(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, projects, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	parsed, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, errors.New("invalid project id"))
		return
	}

	project, err := h.projectService.GetById(r.Context(), parsed, userId)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, project, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	parsed, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, errors.New("invalid project id"))
		return
	}

	var request ProjectBody
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

	serviceRequest := UpdateProjectRequest{
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
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, project, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) UpdateColumn(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	parsedProjectID, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, errors.New("invalid project id"))
		return
	}

	parsedColumnID, err := platformhttp.ParseURLUUID(r, "column_id")
	if err != nil {
		httperr.BadRequestResponse(w, errors.New("invalid project column id"))
		return
	}

	var request UpdateProjectColumnBody
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

	column, err := h.projectService.UpdateColumn(r.Context(), UpdateProjectColumnRequest{
		ProjectId:    parsedProjectID,
		ColumnId:     parsedColumnID,
		UserId:       userId,
		Name:         request.Name,
		Description:  request.Description,
		Color:        request.Color,
		IsDoneColumn: request.IsDoneColumn,
	})
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, column, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func mapProjectColumnsRequest(columns []ProjectColumnBody) []domain.ProjectColumn {
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

func mapDeletedProjectColumnsRequest(columns []DeletedProjectColumnBody) []DeletedProjectColumnRequest {
	mapped := make([]DeletedProjectColumnRequest, len(columns))
	for i, column := range columns {
		mapped[i] = DeletedProjectColumnRequest{
			Id:                  column.Id,
			MoveTasksToColumnId: column.MoveTasksToColumnId,
		}
	}

	return mapped
}

func (h *ProjectHandler) CreateMember(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	parsed, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, errors.New("invalid project id"))
		return
	}

	var request CreateMemberBody
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

	serviceRequest := CreateMemberRequest{
		Email:         request.Email,
		ProjectId:     parsed,
		RequestUserId: userId,
	}

	member, err := h.projectService.CreateMember(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusCreated, member, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	parsed, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, errors.New("invalid project id"))
		return
	}

	parsedMemberId, err := platformhttp.ParseURLUUID(r, "member_id")
	if err != nil {
		httperr.BadRequestResponse(w, errors.New("invalid member id"))
		return
	}

	serviceRequest := RemoveMemberRequest{
		ProjectId:     parsed,
		UserId:        parsedMemberId,
		RequestUserId: userId,
	}

	err = h.projectService.RemoveMember(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusNoContent, nil, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	parsed, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, errors.New("invalid project id"))
		return
	}

	serviceRequest := DeleteProjectRequest{
		ProjectId:     parsed,
		RequestUserId: userId,
	}

	err = h.projectService.Delete(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusNoContent, nil, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) ListActivitiesByProject(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	parsed, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, errors.New("invalid project id"))
		return
	}

	v := validator.New()
	limit, limitErr := platformhttp.ParseLimit(r, 10, 50)
	if limitErr != nil {
		v.Add("limit", limitErr.Error())
	}

	beforeTime := time.Now()
	parsedBeforeTime, parseErr := platformhttp.ParseRFC3339Cursor(r, "before")
	if parseErr != nil {
		v.Add("before", "invalid before date")
	} else if parsedBeforeTime != nil {
		beforeTime = *parsedBeforeTime
	}

	beforeIdUUID := uuid.Nil
	parsedBeforeId, parseErr := platformhttp.ParseOptionalQueryUUID(r, "id")
	if parseErr != nil {
		v.Add("id", "invalid before id")
	} else if parsedBeforeId != nil {
		beforeIdUUID = *parsedBeforeId
	}

	if !v.Valid() {
		httperr.ValidationFailedResponse(w, v)
		return
	}

	serviceRequest := ListActivitiesByProjectRequest{
		ProjectId:       parsed,
		UserId:          userId,
		BeforeCreatedAt: beforeTime,
		BeforeId:        beforeIdUUID,
		Limit:           int32(limit),
	}

	activities, err := h.projectService.ListActivitiesByProject(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, activities, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) ListUsersProjectActivities(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	v := validator.New()
	limit, limitErr := platformhttp.ParseLimit(r, 10, 50)
	if limitErr != nil {
		v.Add("limit", limitErr.Error())
	}

	beforeTime := time.Now()
	parsedBeforeTime, parseErr := platformhttp.ParseRFC3339Cursor(r, "before")
	if parseErr != nil {
		v.Add("before", "invalid before date")
	} else if parsedBeforeTime != nil {
		beforeTime = *parsedBeforeTime
	}

	beforeIdUUID := uuid.Nil
	parsedBeforeId, parseErr := platformhttp.ParseOptionalQueryUUID(r, "id")
	if parseErr != nil {
		v.Add("id", "invalid before id")
	} else if parsedBeforeId != nil {
		beforeIdUUID = *parsedBeforeId
	}

	if !v.Valid() {
		httperr.ValidationFailedResponse(w, v)
		return
	}

	serviceRequest := ListUsersProjectActivitiesRequest{
		UserId:          userId,
		BeforeCreatedAt: beforeTime,
		BeforeId:        beforeIdUUID,
		Limit:           int32(limit),
	}

	activities, err := h.projectService.ListUsersProjectActivities(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, activities, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}

func (h *ProjectHandler) ListMembersByProjectId(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())

	parsed, err := platformhttp.ParseURLUUID(r, "id")
	if err != nil {
		httperr.BadRequestResponse(w, errors.New("invalid project id"))
		return
	}

	members, err := h.projectService.ListMembersByProjectId(r.Context(), parsed, userId)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	err = platformhttp.WriteJSON(w, http.StatusOK, members, nil)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}
}
