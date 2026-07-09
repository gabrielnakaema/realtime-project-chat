package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectRepository struct {
	pool *pgxpool.Pool
}

func optionalText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: value,
		Valid:  true,
	}
}

func NewProjectRepository(pool *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{
		pool: pool,
	}
}

func (pr *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	tx, err := pr.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := queries.New(pr.pool)
	qtx := q.WithTx(tx)

	params := queries.CreateProjectParams{
		UserID:           project.UserId,
		Name:             project.Name,
		Description:      project.Description,
		RepositoryUrl:    optionalText(project.RepositoryURL),
		RepositoryOwner:  optionalText(project.RepositoryOwner),
		RepositoryName:   optionalText(project.RepositoryName),
		DefaultBranch:    optionalText(project.DefaultBranch),
		BranchNamePrefix: optionalText(project.BranchNamePrefix),
	}

	id, err := qtx.CreateProject(ctx, params)
	if err != nil {
		return err
	}

	project.Id = id

	for i, member := range project.Members {
		params := queries.CreateProjectMemberParams{
			UserID:    member.UserId,
			ProjectID: project.Id,
			Role:      string(member.Role),
		}

		id, err := qtx.CreateProjectMember(ctx, params)
		if err != nil {
			return err
		}

		project.Members[i].Id = id
		project.Members[i].ProjectId = project.Id
	}

	for i, column := range project.Columns {
		statusID, err := qtx.CreateProjectColumn(ctx, queries.CreateProjectColumnParams{
			ProjectID:    project.Id,
			Name:         column.Name,
			Description:  column.Description,
			Color:        column.Color,
			Position:     int32(column.Position),
			IsDoneColumn: column.IsDoneColumn,
		})
		if err != nil {
			return err
		}

		project.Columns[i].Id = statusID
		project.Columns[i].ProjectId = project.Id
	}

	return tx.Commit(ctx)
}

func (pr *ProjectRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	q := queries.New(pr.pool)

	projectResult, err := q.GetProjectById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFoundError("project not found")
		}
		return nil, err
	}

	project := domain.Project{
		Id:               projectResult.ID,
		UserId:           projectResult.UserID,
		Name:             projectResult.Name,
		Description:      projectResult.Description,
		CreatedAt:        projectResult.CreatedAt.Time,
		UpdatedAt:        projectResult.UpdatedAt.Time,
		RepositoryURL:    projectResult.RepositoryUrl.String,
		RepositoryOwner:  projectResult.RepositoryOwner.String,
		RepositoryName:   projectResult.RepositoryName.String,
		DefaultBranch:    projectResult.DefaultBranch.String,
		BranchNamePrefix: projectResult.BranchNamePrefix.String,
	}

	bytes, err := json.Marshal(projectResult.Members)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &project.Members)
	if err != nil {
		return nil, err
	}

	bytes, err = json.Marshal(projectResult.Columns)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &project.Columns)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (pr *ProjectRepository) ListByUserId(ctx context.Context, userId uuid.UUID, memberRole string, searchQuery string) ([]domain.Project, error) {
	q := queries.New(pr.pool)

	params := queries.ListProjectsByUserIdParams{
		UserID: userId,
	}

	if memberRole != "" {
		params.Role = pgtype.Text{String: memberRole, Valid: true}
	}

	if searchQuery != "" {
		params.Query = pgtype.Text{String: searchQuery, Valid: true}
	}

	projectResults, err := q.ListProjectsByUserId(ctx, params)
	if err != nil {
		return nil, err
	}

	projects := make([]domain.Project, len(projectResults))

	for i, projectResult := range projectResults {
		projects[i] = domain.Project{
			Id:               projectResult.ID,
			UserId:           projectResult.UserID,
			Name:             projectResult.Name,
			Description:      projectResult.Description,
			CreatedAt:        projectResult.CreatedAt.Time,
			UpdatedAt:        projectResult.UpdatedAt.Time,
			RepositoryURL:    projectResult.RepositoryUrl.String,
			RepositoryOwner:  projectResult.RepositoryOwner.String,
			RepositoryName:   projectResult.RepositoryName.String,
			DefaultBranch:    projectResult.DefaultBranch.String,
			BranchNamePrefix: projectResult.BranchNamePrefix.String,
		}

		bytes, err := json.Marshal(projectResult.Members)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, &projects[i].Members)
		if err != nil {
			return nil, err
		}

		bytes, err = json.Marshal(projectResult.Columns)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, &projects[i].Columns)
		if err != nil {
			return nil, err
		}
	}

	return projects, nil
}

func (pr *ProjectRepository) Update(ctx context.Context, project *domain.Project) error {

	q := queries.New(pr.pool)

	params := queries.UpdateProjectParams{
		Name:             project.Name,
		Description:      project.Description,
		RepositoryUrl:    optionalText(project.RepositoryURL),
		RepositoryOwner:  optionalText(project.RepositoryOwner),
		RepositoryName:   optionalText(project.RepositoryName),
		DefaultBranch:    optionalText(project.DefaultBranch),
		BranchNamePrefix: optionalText(project.BranchNamePrefix),
		ID:               project.Id,
	}

	err := q.UpdateProject(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFoundError("project not found")
		}
		return err
	}

	return nil
}

type ProjectColumnDeletion struct {
	ColumnID     uuid.UUID
	TargetColumn domain.ProjectColumn
}

type UpdateProjectWithColumnsParams struct {
	Project   *domain.Project
	Columns   []domain.ProjectColumn
	Deletions []ProjectColumnDeletion
}

func (pr *ProjectRepository) UpdateWithColumns(ctx context.Context, params UpdateProjectWithColumnsParams) error {
	tx, err := pr.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := queries.New(pr.pool)
	qtx := q.WithTx(tx)

	project := params.Project

	err = qtx.UpdateProject(ctx, queries.UpdateProjectParams{
		Name:             project.Name,
		Description:      project.Description,
		RepositoryUrl:    optionalText(project.RepositoryURL),
		RepositoryOwner:  optionalText(project.RepositoryOwner),
		RepositoryName:   optionalText(project.RepositoryName),
		DefaultBranch:    optionalText(project.DefaultBranch),
		BranchNamePrefix: optionalText(project.BranchNamePrefix),
		ID:               project.Id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotFoundError("project not found")
		}
		return err
	}

	for i, column := range params.Columns {
		if column.Id == uuid.Nil {
			continue
		}

		if err := qtx.UpdateProjectColumn(ctx, queries.UpdateProjectColumnParams{
			Name:         column.Name,
			Description:  column.Description,
			Color:        column.Color,
			Position:     int32(i + 1000),
			IsDoneColumn: false,
			ID:           column.Id,
			ProjectID:    project.Id,
		}); err != nil {
			return err
		}
	}

	for i := range params.Columns {
		params.Columns[i].Position = i
		params.Columns[i].ProjectId = project.Id

		if params.Columns[i].Id == uuid.Nil {
			id, err := qtx.CreateProjectColumn(ctx, queries.CreateProjectColumnParams{
				ProjectID:    project.Id,
				Name:         params.Columns[i].Name,
				Description:  params.Columns[i].Description,
				Color:        params.Columns[i].Color,
				Position:     int32(params.Columns[i].Position),
				IsDoneColumn: params.Columns[i].IsDoneColumn,
			})
			if err != nil {
				return err
			}

			params.Columns[i].Id = id
			continue
		}

		if err := qtx.UpdateProjectColumn(ctx, queries.UpdateProjectColumnParams{
			Name:         params.Columns[i].Name,
			Description:  params.Columns[i].Description,
			Color:        params.Columns[i].Color,
			Position:     int32(params.Columns[i].Position),
			IsDoneColumn: params.Columns[i].IsDoneColumn,
			ID:           params.Columns[i].Id,
			ProjectID:    project.Id,
		}); err != nil {
			return err
		}
	}

	for _, del := range params.Deletions {
		if err := qtx.ReassignTasksToProjectColumn(ctx, queries.ReassignTasksToProjectColumnParams{
			ProjectColumnID:   del.TargetColumn.Id,
			Column2:           del.TargetColumn.IsDoneColumn,
			ProjectColumnID_2: del.ColumnID,
		}); err != nil {
			return err
		}

		if err := qtx.DeleteProjectColumn(ctx, queries.DeleteProjectColumnParams{
			ID:        del.ColumnID,
			ProjectID: project.Id,
		}); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (pr *ProjectRepository) CreateMember(ctx context.Context, member *domain.ProjectMember) error {
	q := queries.New(pr.pool)

	params := queries.CreateProjectMemberParams{
		UserID:    member.UserId,
		ProjectID: member.ProjectId,
		Role:      string(member.Role),
	}

	id, err := q.CreateProjectMember(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return domain.DuplicateEntryError("member already exists")
			}
			return err
		}

		return err
	}

	member.Id = id

	return nil
}

func (pr *ProjectRepository) RemoveMember(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) error {
	tx, err := pr.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := queries.New(pr.pool)
	qtx := q.WithTx(tx)

	params := queries.RemoveProjectMemberParams{
		ProjectID: projectId,
		UserID:    userId,
	}

	err = qtx.RemoveProjectMember(ctx, params)
	if err != nil {
		return err
	}

	err = qtx.MarkProjectUpdatedAt(ctx, projectId)
	if err != nil {
		return err
	}

	removeTasksResponsibleParams := queries.RemoveProjectTasksResponsibleByUserIdParams{
		ResponsibleID: pgtype.UUID{
			Bytes: userId,
			Valid: true,
		},
		ProjectID: projectId,
	}

	err = qtx.RemoveProjectTasksResponsibleByUserId(ctx, removeTasksResponsibleParams)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (pr *ProjectRepository) GetMemberByUserIdAndProjectId(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) (*domain.ProjectMember, error) {
	q := queries.New(pr.pool)

	params := queries.GetProjectMemberByUserIdAndProjectIdParams{
		ProjectID: projectId,
		UserID:    userId,
	}

	memberResult, err := q.GetProjectMemberByUserIdAndProjectId(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFoundError("member not found")
		}
		return nil, err
	}

	member := domain.ProjectMember{
		Id:        memberResult.ID,
		UserId:    memberResult.UserID,
		ProjectId: memberResult.ProjectID,
		Role:      domain.ProjectMemberRole(memberResult.Role),
	}

	return &member, nil
}

func (pr *ProjectRepository) MarkUpdatedAt(ctx context.Context, projectId uuid.UUID) error {
	q := queries.New(pr.pool)

	return q.MarkProjectUpdatedAt(ctx, projectId)
}

func (pr *ProjectRepository) CreateColumn(ctx context.Context, status *domain.ProjectColumn) error {
	q := queries.New(pr.pool)

	id, err := q.CreateProjectColumn(ctx, queries.CreateProjectColumnParams{
		ProjectID:    status.ProjectId,
		Name:         status.Name,
		Description:  status.Description,
		Color:        status.Color,
		Position:     int32(status.Position),
		IsDoneColumn: status.IsDoneColumn,
	})
	if err != nil {
		return err
	}

	status.Id = id
	return nil
}

func (pr *ProjectRepository) UpdateColumn(ctx context.Context, status *domain.ProjectColumn) error {
	q := queries.New(pr.pool)

	return q.UpdateProjectColumn(ctx, queries.UpdateProjectColumnParams{
		Name:         status.Name,
		Description:  status.Description,
		Color:        status.Color,
		Position:     int32(status.Position),
		IsDoneColumn: status.IsDoneColumn,
		ID:           status.Id,
		ProjectID:    status.ProjectId,
	})
}

func (pr *ProjectRepository) DeleteColumn(ctx context.Context, projectId uuid.UUID, statusId uuid.UUID) error {
	q := queries.New(pr.pool)

	return q.DeleteProjectColumn(ctx, queries.DeleteProjectColumnParams{
		ID:        statusId,
		ProjectID: projectId,
	})
}

func (pr *ProjectRepository) ReassignTasksToColumn(ctx context.Context, fromStatusId uuid.UUID, toStatus *domain.ProjectColumn) error {
	q := queries.New(pr.pool)

	return q.ReassignTasksToProjectColumn(ctx, queries.ReassignTasksToProjectColumnParams{
		ProjectColumnID:   toStatus.Id,
		Column2:           toStatus.IsDoneColumn,
		ProjectColumnID_2: fromStatusId,
	})
}

func (pr *ProjectRepository) GetColumnById(ctx context.Context, id uuid.UUID) (*domain.ProjectColumn, error) {
	q := queries.New(pr.pool)

	status, err := q.GetProjectColumnById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFoundError("project status not found")
		}
		return nil, err
	}

	return &domain.ProjectColumn{
		Id:           status.ID,
		ProjectId:    status.ProjectID,
		Name:         status.Name,
		Description:  status.Description,
		Color:        status.Color,
		Position:     int(status.Position),
		IsDoneColumn: status.IsDoneColumn,
		CreatedAt:    status.CreatedAt.Time,
		UpdatedAt:    status.UpdatedAt.Time,
	}, nil
}

func (pr *ProjectRepository) ListMembersByProjectId(ctx context.Context, projectId uuid.UUID) ([]domain.ProjectMember, error) {
	q := queries.New(pr.pool)

	membersResult, err := q.ListProjectMembersByProjectId(ctx, projectId)
	if err != nil {
		return nil, err
	}

	members := make([]domain.ProjectMember, len(membersResult))

	for i, member := range membersResult {
		members[i] = domain.ProjectMember{
			Id:        member.ID,
			UserId:    member.UserID,
			ProjectId: member.ProjectMemberProjectID,
			Role:      domain.ProjectMemberRole(member.ProjectMemberRole),
			User: &domain.User{
				Id:        member.UserID,
				Name:      member.UserName,
				Email:     member.UserEmail,
				CreatedAt: member.UserCreatedAt.Time,
			},
		}
	}

	return members, nil
}
