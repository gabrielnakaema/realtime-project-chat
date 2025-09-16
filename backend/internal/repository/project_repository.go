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
		UserID:      project.UserId,
		Name:        project.Name,
		Description: project.Description,
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
		Id:          projectResult.ID,
		UserId:      projectResult.UserID,
		Name:        projectResult.Name,
		Description: projectResult.Description,
		CreatedAt:   projectResult.CreatedAt.Time,
		UpdatedAt:   projectResult.UpdatedAt.Time,
	}

	bytes, err := json.Marshal(projectResult.Members)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &project.Members)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (pr *ProjectRepository) ListByUserId(ctx context.Context, userId uuid.UUID, memberRole string) ([]domain.Project, error) {
	q := queries.New(pr.pool)

	params := queries.ListProjectsByUserIdParams{
		UserID: userId,
	}

	if memberRole != "" {
		params.Role = pgtype.Text{String: memberRole, Valid: true}
	}

	projectResults, err := q.ListProjectsByUserId(ctx, params)
	if err != nil {
		return nil, err
	}

	projects := make([]domain.Project, len(projectResults))

	for i, projectResult := range projectResults {
		projects[i] = domain.Project{
			Id:          projectResult.ID,
			UserId:      projectResult.UserID,
			Name:        projectResult.Name,
			Description: projectResult.Description,
			CreatedAt:   projectResult.CreatedAt.Time,
			UpdatedAt:   projectResult.UpdatedAt.Time,
		}

		bytes, err := json.Marshal(projectResult.Members)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, &projects[i].Members)
		if err != nil {
			return nil, err
		}
	}

	return projects, nil
}

func (pr *ProjectRepository) Update(ctx context.Context, project *domain.Project) error {

	q := queries.New(pr.pool)

	params := queries.UpdateProjectParams{
		Name:        project.Name,
		Description: project.Description,
		ID:          project.Id,
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
	q := queries.New(pr.pool)

	params := queries.RemoveProjectMemberParams{
		ProjectID: projectId,
		UserID:    userId,
	}

	err := q.RemoveProjectMember(ctx, params)
	if err != nil {
		return err
	}

	return nil
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
