package repository

import (
	"context"
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

func (r *TaskCommentRepository) ListByTaskID(ctx context.Context, taskID uuid.UUID, before time.Time, beforeID uuid.UUID, limit int) (*utils.CursorPaginated[domain.TaskComment], error) {
	q := queries.New(r.pool)

	rows, err := q.ListTaskComments(ctx, queries.ListTaskCommentsParams{
		TaskID: taskID,
		Limit:  int32(limit + 1),
		Column3: pgtype.Timestamptz{
			Time:  before,
			Valid: true,
		},
		Column4: beforeID,
	})
	if err != nil {
		return nil, err
	}

	nodesByID := make(map[uuid.UUID]*taskCommentNode, len(rows))
	pendingChildren := make(map[uuid.UUID][]*taskCommentNode)
	rootNodes := make([]*taskCommentNode, 0)
	rootCount := 0
	hasNext := false

	for _, row := range rows {
		if row.Level == 0 {
			rootCount++
			if rootCount > limit {
				hasNext = true
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

	comments := make([]domain.TaskComment, 0, len(rootNodes))
	for _, root := range rootNodes {
		comments = append(comments, buildTaskCommentTree(root))
	}

	return &utils.CursorPaginated[domain.TaskComment]{
		Data:    comments,
		HasNext: hasNext,
	}, nil
}

func buildTaskCommentTree(node *taskCommentNode) domain.TaskComment {
	comment := node.comment
	comment.Replies = make([]domain.TaskComment, 0, len(node.replies))

	for _, reply := range node.replies {
		comment.Replies = append(comment.Replies, buildTaskCommentTree(reply))
	}

	return comment
}
