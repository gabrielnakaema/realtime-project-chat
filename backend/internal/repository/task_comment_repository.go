package repository

import (
	"context"
	"errors"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/queries"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskCommentRepository struct {
	pool *pgxpool.Pool
}

type taskCommentNode struct {
	comment domain.TaskComment
	replies []*taskCommentNode
}

func NewTaskCommentRepository(pool *pgxpool.Pool) *TaskCommentRepository {
	return &TaskCommentRepository{
		pool: pool,
	}
}

func (r *TaskCommentRepository) Create(ctx context.Context, comment *domain.TaskComment, parentCommentID *uuid.UUID) error {
	q := queries.New(r.pool)

	params := queries.CreateTaskCommentParams{
		TaskID:  comment.Task.Id,
		UserID:  comment.User.Id,
		Content: comment.Content,
	}

	if parentCommentID != nil && *parentCommentID != uuid.Nil {
		params.ParentCommentID = pgtype.UUID{
			Bytes: *parentCommentID,
			Valid: true,
		}
		parentID := parentCommentID.String()
		comment.ParentCommentID = &parentID
	}

	id, err := q.CreateTaskComment(ctx, params)
	if err != nil {
		return err
	}

	comment.ID = id.String()

	return nil
}

func (r *TaskCommentRepository) ListByTaskID(
	ctx context.Context,
	taskID uuid.UUID,
	before *time.Time,
	beforeID *uuid.UUID,
	after *time.Time,
	afterID *uuid.UUID,
	limit int,
) (*utils.CursorPaginated[domain.TaskComment], error) {
	q := queries.New(r.pool)

	if before != nil && after != nil {
		return nil, errors.New("before and after cannot be used together")
	}

	paginationMode := "before"
	if after != nil {
		paginationMode = "after"
		afterRows, afterErr := q.ListTaskCommentsAfter(ctx, queries.ListTaskCommentsAfterParams{
			TaskID: taskID,
			Limit:  int32(limit + 1),
			Column3: pgtype.Timestamptz{
				Time:  *after,
				Valid: true,
			},
			Column4: cursorUUID(afterID),
		})
		if afterErr != nil {
			return nil, afterErr
		}

		return r.buildPaginatedComments(ctx, taskID, afterRows, limit, paginationMode)
	}

	if before == nil {
		now := time.Now()
		before = &now
	}

	rows, err := q.ListTaskComments(ctx, queries.ListTaskCommentsParams{
		TaskID: taskID,
		Limit:  int32(limit + 1),
		Column3: pgtype.Timestamptz{
			Time:  *before,
			Valid: true,
		},
		Column4: cursorUUID(beforeID),
	})
	if err != nil {
		return nil, err
	}

	return r.buildPaginatedComments(ctx, taskID, rows, limit, paginationMode)
}

func (r *TaskCommentRepository) buildPaginatedComments(
	ctx context.Context,
	taskID uuid.UUID,
	rows interface{},
	limit int,
	paginationMode string,
) (*utils.CursorPaginated[domain.TaskComment], error) {
	normalizedRows := normalizeTaskCommentRows(rows)
	nodesByID := make(map[uuid.UUID]*taskCommentNode, len(normalizedRows))
	pendingChildren := make(map[uuid.UUID][]*taskCommentNode)
	rootNodes := make([]*taskCommentNode, 0)
	rootCount := 0
	hasNext := false
	hasPrevious := false

	for _, row := range normalizedRows {
		if row.Level == 0 {
			rootCount++
			if rootCount > limit {
				if paginationMode == "after" {
					hasPrevious = true
				} else {
					hasNext = true
				}
				break
			}
		}

		node := &taskCommentNode{
			comment: domain.TaskComment{
				ID:        row.ID.String(),
				Content:   row.Content,
				CreatedAt: row.CreatedAt.Time,
				UpdatedAt: row.UpdatedAt.Time,
				User: &domain.User{
					Id:        row.CommentUserID,
					Name:      row.CommentUserName,
					Email:     row.CommentUserEmail,
					CreatedAt: row.CommentUserCreatedAt.Time,
				},
				Replies: []domain.TaskComment{},
			},
			replies: []*taskCommentNode{},
		}

		if row.ParentCommentID.Valid {
			parentID := uuid.UUID(row.ParentCommentID.Bytes).String()
			node.comment.ParentCommentID = &parentID
		}

		nodesByID[row.ID] = node

		if children, ok := pendingChildren[row.ID]; ok {
			node.replies = append(node.replies, children...)
			delete(pendingChildren, row.ID)
		}

		if row.ParentCommentID.Valid {
			parentID := row.ParentCommentID.Bytes
			if parentNode, ok := nodesByID[parentID]; ok {
				parentNode.replies = append(parentNode.replies, node)
			} else {
				pendingChildren[parentID] = append(pendingChildren[parentID], node)
			}
			continue
		}

		rootNodes = append(rootNodes, node)
	}

	if paginationMode == "after" {
		reverseTaskCommentNodes(rootNodes)
	}

	comments := make([]domain.TaskComment, 0, len(rootNodes))
	for _, root := range rootNodes {
		comments = append(comments, buildTaskCommentTree(root))
	}

	if len(comments) > 0 {
		if paginationMode != "after" {
			var err error
			hasPrevious, err = r.hasRootCommentsAfter(ctx, taskID, comments[0])
			if err != nil {
				return nil, err
			}
		}

		if paginationMode != "before" {
			var err error
			hasNext, err = r.hasRootCommentsBefore(ctx, taskID, comments[len(comments)-1])
			if err != nil {
				return nil, err
			}
		}
	}

	return &utils.CursorPaginated[domain.TaskComment]{
		Data:        comments,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}, nil
}

type normalizedTaskCommentRow struct {
	ID                   uuid.UUID
	TaskID               uuid.UUID
	UserID               uuid.UUID
	Content              string
	ParentCommentID      pgtype.UUID
	CreatedAt            pgtype.Timestamptz
	UpdatedAt            pgtype.Timestamptz
	Level                int32
	CommentUserID        uuid.UUID
	CommentUserName      string
	CommentUserEmail     string
	CommentUserCreatedAt pgtype.Timestamptz
}

func normalizeTaskCommentRows(rows interface{}) []normalizedTaskCommentRow {
	switch typedRows := rows.(type) {
	case []queries.ListTaskCommentsRow:
		normalized := make([]normalizedTaskCommentRow, 0, len(typedRows))
		for _, row := range typedRows {
			normalized = append(normalized, normalizedTaskCommentRow{
				ID:                   row.ID,
				TaskID:               row.TaskID,
				UserID:               row.UserID,
				Content:              row.Content,
				ParentCommentID:      row.ParentCommentID,
				CreatedAt:            row.CreatedAt,
				UpdatedAt:            row.UpdatedAt,
				Level:                row.Level,
				CommentUserID:        row.CommentUserID,
				CommentUserName:      row.CommentUserName,
				CommentUserEmail:     row.CommentUserEmail,
				CommentUserCreatedAt: row.CommentUserCreatedAt,
			})
		}
		return normalized
	case []queries.ListTaskCommentsAfterRow:
		normalized := make([]normalizedTaskCommentRow, 0, len(typedRows))
		for _, row := range typedRows {
			normalized = append(normalized, normalizedTaskCommentRow{
				ID:                   row.ID,
				TaskID:               row.TaskID,
				UserID:               row.UserID,
				Content:              row.Content,
				ParentCommentID:      row.ParentCommentID,
				CreatedAt:            row.CreatedAt,
				UpdatedAt:            row.UpdatedAt,
				Level:                row.Level,
				CommentUserID:        row.CommentUserID,
				CommentUserName:      row.CommentUserName,
				CommentUserEmail:     row.CommentUserEmail,
				CommentUserCreatedAt: row.CommentUserCreatedAt,
			})
		}
		return normalized
	default:
		return []normalizedTaskCommentRow{}
	}
}

func (r *TaskCommentRepository) hasRootCommentsAfter(ctx context.Context, taskID uuid.UUID, comment domain.TaskComment) (bool, error) {
	id, err := uuid.Parse(comment.ID)
	if err != nil {
		return false, err
	}

	var exists bool
	err = r.pool.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM task_comments
			WHERE task_id = $1
			  AND parent_comment_id IS NULL
			  AND (created_at, id) > ($2::timestamptz, $3::uuid)
		)
		`,
		taskID,
		comment.CreatedAt,
		id,
	).Scan(&exists)
	return exists, err
}

func (r *TaskCommentRepository) hasRootCommentsBefore(ctx context.Context, taskID uuid.UUID, comment domain.TaskComment) (bool, error) {
	id, err := uuid.Parse(comment.ID)
	if err != nil {
		return false, err
	}

	var exists bool
	err = r.pool.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM task_comments
			WHERE task_id = $1
			  AND parent_comment_id IS NULL
			  AND (created_at, id) < ($2::timestamptz, $3::uuid)
		)
		`,
		taskID,
		comment.CreatedAt,
		id,
	).Scan(&exists)
	return exists, err
}

func cursorUUID(id *uuid.UUID) uuid.UUID {
	if id == nil {
		return uuid.Nil
	}

	return *id
}

func reverseTaskCommentNodes(nodes []*taskCommentNode) {
	for left, right := 0, len(nodes)-1; left < right; left, right = left+1, right-1 {
		nodes[left], nodes[right] = nodes[right], nodes[left]
	}
}

func buildTaskCommentTree(node *taskCommentNode) domain.TaskComment {
	comment := node.comment
	comment.Replies = make([]domain.TaskComment, 0, len(node.replies))

	for _, reply := range node.replies {
		comment.Replies = append(comment.Replies, buildTaskCommentTree(reply))
	}

	return comment
}
