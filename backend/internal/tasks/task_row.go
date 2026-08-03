package tasks

import (
	"encoding/json"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	taskqueries "github.com/gabrielnakaema/project-chat/internal/tasks/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type taskRecord struct {
	id              uuid.UUID
	projectID       uuid.UUID
	authorID        uuid.UUID
	title           string
	description     string
	code            pgtype.Text
	projectColumnID uuid.UUID
	priority        string
	order           string
	version         int32
	responsibleID   pgtype.UUID
	dueDate         pgtype.Timestamptz
	doneAt          pgtype.Timestamptz
	archivedAt      pgtype.Timestamptz
	createdAt       pgtype.Timestamptz
	updatedAt       pgtype.Timestamptz
	tags            any
	dependsOnIDs    any

	project       *domain.Project
	projectColumn *domain.ProjectColumn
	author        *domain.User
	responsible   *domain.User
}

func normalizeListByProjectTask(row taskqueries.ListTasksByProjectIdRow) taskRecord {
	record := taskRecord{
		id:              row.ID,
		projectID:       row.ProjectID,
		authorID:        row.AuthorID,
		title:           row.Title,
		description:     row.Description,
		code:            row.Code,
		projectColumnID: row.ProjectColumnID,
		priority:        row.Priority,
		order:           row.TaskOrder,
		version:         row.Version,
		responsibleID:   row.ResponsibleID,
		dueDate:         row.DueDate,
		doneAt:          row.DoneAt,
		archivedAt:      row.ArchivedAt,
		createdAt:       row.CreatedAt,
		updatedAt:       row.UpdatedAt,
		tags:            row.Tags,
		dependsOnIDs:    row.DependsOnTaskIds,
		projectColumn: mapProjectColumn(
			row.ProjectColumnID2,
			row.ProjectColumnProjectID,
			row.ProjectColumnName,
			row.ProjectColumnDescription,
			row.ProjectColumnColor,
			row.ProjectColumnPosition,
			row.ProjectColumnIsDoneColumn,
			row.ProjectColumnCreatedAt,
			row.ProjectColumnUpdatedAt,
		),
	}

	if row.AuthorAuthorID.Valid {
		record.author = &domain.User{
			Id:        uuid.UUID(row.AuthorAuthorID.Bytes),
			Name:      row.AuthorName.String,
			Email:     row.AuthorEmail.String,
			CreatedAt: row.AuthorCreatedAt.Time,
		}
	}
	if row.ResponsibleResponsibleID.Valid {
		record.responsible = &domain.User{
			Id:        uuid.UUID(row.ResponsibleResponsibleID.Bytes),
			Name:      row.ResponsibleName.String,
			Email:     row.ResponsibleEmail.String,
			CreatedAt: row.ResponsibleCreatedAt.Time,
		}
	}

	return record
}

func normalizeUserDueTask(row taskqueries.ListUserDueTasksRow) taskRecord {
	record := taskRecord{
		id:              row.ID,
		projectID:       row.ProjectID,
		authorID:        row.AuthorID,
		title:           row.Title,
		description:     row.Description,
		code:            row.Code,
		projectColumnID: row.ProjectColumnID,
		priority:        row.Priority,
		order:           row.TaskOrder,
		version:         row.Version,
		responsibleID:   row.ResponsibleID,
		dueDate:         row.DueDate,
		doneAt:          row.DoneAt,
		archivedAt:      row.ArchivedAt,
		createdAt:       row.CreatedAt,
		updatedAt:       row.UpdatedAt,
		tags:            row.Tags,
		dependsOnIDs:    row.DependsOnTaskIds,
		projectColumn: mapProjectColumn(
			row.ProjectColumnID2,
			row.ProjectColumnProjectID,
			row.ProjectColumnName,
			row.ProjectColumnDescription,
			row.ProjectColumnColor,
			row.ProjectColumnPosition,
			row.ProjectColumnIsDoneColumn,
			row.ProjectColumnCreatedAt,
			row.ProjectColumnUpdatedAt,
		),
		project: &domain.Project{
			Id:               row.ProjectProjectID,
			Name:             row.ProjectName,
			Description:      row.ProjectDescription,
			RepositoryURL:    row.ProjectRepositoryUrl.String,
			RepositoryOwner:  row.ProjectRepositoryOwner.String,
			RepositoryName:   row.ProjectRepositoryName.String,
			DefaultBranch:    row.ProjectDefaultBranch.String,
			BranchNamePrefix: row.ProjectBranchNamePrefix.String,
			CreatedAt:        row.ProjectCreatedAt.Time,
			UpdatedAt:        row.ProjectUpdatedAt.Time,
			UserId:           row.ProjectUserID,
		},
	}

	if row.AuthorAuthorID.Valid {
		record.author = &domain.User{
			Id:        uuid.UUID(row.AuthorAuthorID.Bytes),
			Name:      row.AuthorName.String,
			Email:     row.AuthorEmail.String,
			CreatedAt: row.AuthorCreatedAt.Time,
		}
	}
	if row.ResponsibleResponsibleID.Valid {
		record.responsible = &domain.User{
			Id:        uuid.UUID(row.ResponsibleResponsibleID.Bytes),
			Name:      row.ResponsibleName.String,
			Email:     row.ResponsibleEmail.String,
			CreatedAt: row.ResponsibleCreatedAt.Time,
		}
	}

	return record
}

func normalizeSearchTask(row taskqueries.SearchTasksForUserRow) taskRecord {
	record := taskRecord{
		id:              row.ID,
		projectID:       row.ProjectID,
		authorID:        row.AuthorID,
		title:           row.Title,
		description:     row.Description,
		code:            row.Code,
		projectColumnID: row.ProjectColumnID,
		priority:        row.Priority,
		order:           row.TaskOrder,
		version:         row.Version,
		responsibleID:   row.ResponsibleID,
		dueDate:         row.DueDate,
		doneAt:          row.DoneAt,
		archivedAt:      row.ArchivedAt,
		createdAt:       row.CreatedAt,
		updatedAt:       row.UpdatedAt,
		tags:            row.Tags,
		dependsOnIDs:    row.DependsOnTaskIds,
		projectColumn: mapProjectColumn(
			row.ProjectColumnID2,
			row.ProjectColumnProjectID,
			row.ProjectColumnName,
			row.ProjectColumnDescription,
			row.ProjectColumnColor,
			row.ProjectColumnPosition,
			row.ProjectColumnIsDoneColumn,
			row.ProjectColumnCreatedAt,
			row.ProjectColumnUpdatedAt,
		),
		project: &domain.Project{
			Id:               row.ProjectProjectID,
			Name:             row.ProjectName,
			Description:      row.ProjectDescription,
			RepositoryURL:    row.ProjectRepositoryUrl.String,
			RepositoryOwner:  row.ProjectRepositoryOwner.String,
			RepositoryName:   row.ProjectRepositoryName.String,
			DefaultBranch:    row.ProjectDefaultBranch.String,
			BranchNamePrefix: row.ProjectBranchNamePrefix.String,
			CreatedAt:        row.ProjectCreatedAt.Time,
			UpdatedAt:        row.ProjectUpdatedAt.Time,
			UserId:           row.ProjectUserID,
		},
	}

	if row.AuthorAuthorID.Valid {
		record.author = &domain.User{
			Id:        uuid.UUID(row.AuthorAuthorID.Bytes),
			Name:      row.AuthorName.String,
			Email:     row.AuthorEmail.String,
			CreatedAt: row.AuthorCreatedAt.Time,
		}
	}
	if row.ResponsibleResponsibleID.Valid {
		record.responsible = &domain.User{
			Id:        uuid.UUID(row.ResponsibleResponsibleID.Bytes),
			Name:      row.ResponsibleName.String,
			Email:     row.ResponsibleEmail.String,
			CreatedAt: row.ResponsibleCreatedAt.Time,
		}
	}

	return record
}

func toDomainTask(record taskRecord) (domain.Task, error) {
	task := domain.Task{
		Id:              record.id,
		ProjectId:       record.projectID,
		AuthorId:        record.authorID,
		Title:           record.title,
		Description:     record.description,
		ProjectColumnId: record.projectColumnID,
		Priority:        domain.TaskPriority(record.priority),
		Order:           record.order,
		Version:         int(record.version),
		CreatedAt:       record.createdAt.Time,
		UpdatedAt:       record.updatedAt.Time,
		Project:         record.project,
		ProjectColumn:   record.projectColumn,
		Author:          record.author,
		Responsible:     record.responsible,
	}

	if record.code.Valid {
		task.Code = record.code.String
	}
	if record.responsibleID.Valid {
		responsibleID := uuid.UUID(record.responsibleID.Bytes)
		task.ResponsibleId = &responsibleID
	}
	if record.dueDate.Valid {
		dueDate := record.dueDate.Time
		task.DueDate = &dueDate
	}
	if record.doneAt.Valid {
		doneAt := record.doneAt.Time
		task.DoneAt = &doneAt
	}
	if record.archivedAt.Valid {
		archivedAt := record.archivedAt.Time
		task.ArchivedAt = &archivedAt
	}

	if record.tags != nil {
		data, err := json.Marshal(record.tags)
		if err != nil {
			return domain.Task{}, err
		}
		if err := json.Unmarshal(data, &task.Tags); err != nil {
			return domain.Task{}, err
		}
	}

	if err := mapDependsOnTaskIdsFromResult(record.dependsOnIDs, &task); err != nil {
		return domain.Task{}, err
	}

	if task.ProjectColumn != nil {
		task.Status = compatibilityTaskStatus(task.ProjectColumn.Name, task.ArchivedAt)
	}

	return task, nil
}
