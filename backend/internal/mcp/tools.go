package mcp

import (
	"context"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
)

type toolHandlerFunc func(h *Handler, ctx context.Context, principal principal, args map[string]any) (map[string]any, error)

func (h *Handler) callTool(ctx context.Context, principal principal, params toolCallParams) (map[string]any, error) {
	spec, ok := findToolSpec(params.Name)
	if !ok {
		return nil, apperr.NotFoundError("tool not found")
	}

	if err := requireScope(principal, spec.RequiredScope); err != nil {
		return nil, err
	}

	return spec.Handle(h, ctx, principal, params.Arguments)
}

func requireScope(principal principal, scope domain.MCPAPIScope) error {
	if principal.HasScope(scope) {
		return nil
	}

	return apperr.ForbiddenError("missing required scope")
}

func (h *Handler) handleListProjects(ctx context.Context, principal principal, args map[string]any) (map[string]any, error) {
	projects, err := h.projectService.ListByUserID(ctx, ListProjectsRequest{
		UserID: principal.UserID,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"projects": projects}, nil
}

func (h *Handler) handleListProjectBoard(ctx context.Context, principal principal, args map[string]any) (map[string]any, error) {
	projectID, err := requiredUUIDArg(args, "project_id")
	if err != nil {
		return nil, err
	}
	project, err := h.projectService.GetByID(ctx, projectID, principal.UserID)
	if err != nil {
		return nil, err
	}
	projectColumnIDs, err := optionalUUIDSliceArg(args, "project_column_ids")
	if err != nil {
		return nil, err
	}
	limitPerColumn, err := optionalLimitArg(args, "limit_per_column", 15)
	if err != nil {
		return nil, err
	}
	includeArchived, err := optionalBoolArg(args, "include_archived", false)
	if err != nil {
		return nil, err
	}
	grouped, err := h.taskService.GroupByColumn(ctx, GroupByColumnRequest{
		ProjectID:        projectID,
		UserID:           principal.UserID,
		ProjectColumnIDs: projectColumnIDs,
		Archived:         includeArchived,
		Limit:            limitPerColumn,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"project":         project,
		"columns":         project.Columns,
		"tasks_by_column": grouped,
	}, nil
}

func (h *Handler) handleCreateTask(ctx context.Context, principal principal, args map[string]any) (map[string]any, error) {
	projectID, err := requiredUUIDArg(args, "project_id")
	if err != nil {
		return nil, err
	}
	projectColumnID, err := requiredUUIDArg(args, "project_column_id")
	if err != nil {
		return nil, err
	}
	title, err := requiredStringArg(args, "title")
	if err != nil {
		return nil, err
	}
	description, err := requiredStringArg(args, "description")
	if err != nil {
		return nil, err
	}
	priority, err := requiredTaskPriorityArg(args, "priority")
	if err != nil {
		return nil, err
	}
	responsibleID, err := optionalUUIDArg(args, "responsible_id")
	if err != nil {
		return nil, err
	}
	dueDate, err := optionalTimeArg(args, "due_date")
	if err != nil {
		return nil, err
	}
	tags, err := optionalStringSliceArg(args, "tags")
	if err != nil {
		return nil, err
	}
	dependsOnTaskIDs, err := optionalUUIDSliceArg(args, "depends_on_task_ids")
	if err != nil {
		return nil, err
	}
	code, err := optionalTrimmedStringArg(args, "code")
	if err != nil {
		return nil, err
	}
	ctx = domain.WithActionOrigin(ctx, domain.ActionOriginMCPAgent)
	task, err := h.taskService.Create(ctx, CreateTaskRequest{
		ProjectID:        projectID,
		ProjectColumnID:  projectColumnID,
		Title:            title,
		Description:      description,
		Code:             code,
		RequestUserID:    principal.UserID,
		Priority:         string(priority),
		ResponsibleID:    responsibleID,
		DueDate:          dueDate,
		Tags:             tags,
		DependsOnTaskIDs: dependsOnTaskIDs,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"task": task}, nil
}

func (h *Handler) handleGetTask(ctx context.Context, principal principal, args map[string]any) (map[string]any, error) {
	taskID, err := requiredUUIDArg(args, "task_id")
	if err != nil {
		return nil, err
	}
	task, err := h.taskService.GetByID(ctx, taskID, principal.UserID)
	if err != nil {
		return nil, err
	}
	if err := h.attachRequestedComments(ctx, principal, args, task); err != nil {
		return nil, err
	}
	return map[string]any{"task": task}, nil
}

func (h *Handler) handleFindTaskByCode(ctx context.Context, principal principal, args map[string]any) (map[string]any, error) {
	projectID, err := requiredUUIDArg(args, "project_id")
	if err != nil {
		return nil, err
	}
	code, err := requiredStringArg(args, "code")
	if err != nil {
		return nil, err
	}
	task, err := h.taskService.FindTaskByCode(ctx, FindTaskByCodeRequest{
		ProjectID: projectID,
		UserID:    principal.UserID,
		Code:      code,
	})
	if err != nil {
		return nil, err
	}
	if err := h.attachRequestedComments(ctx, principal, args, task); err != nil {
		return nil, err
	}
	return map[string]any{"task": task}, nil
}

func (h *Handler) attachRequestedComments(ctx context.Context, principal principal, args map[string]any, task *domain.Task) error {
	includeComments, err := optionalBoolArg(args, "include_comments", false)
	if err != nil {
		return err
	}
	if !includeComments {
		return nil
	}
	if err := requireScope(principal, domain.MCPAPIScopeTasksCommentsRead); err != nil {
		return err
	}
	commentsLimit, err := optionalLimitArg(args, "comments_limit", 10)
	if err != nil {
		return err
	}

	comments, err := h.taskCommentService.ListByTaskID(ctx, ListTaskCommentsRequest{
		TaskID:        task.Id,
		RequestUserID: principal.UserID,
		Limit:         commentsLimit,
	})
	if err != nil {
		return err
	}
	task.Comments = comments.Data
	return nil
}

func (h *Handler) handleListTaskComments(ctx context.Context, principal principal, args map[string]any) (map[string]any, error) {
	taskID, err := requiredUUIDArg(args, "task_id")
	if err != nil {
		return nil, err
	}
	commentsLimit, err := optionalLimitArg(args, "limit", 10)
	if err != nil {
		return nil, err
	}
	comments, err := h.taskCommentService.ListByTaskID(ctx, ListTaskCommentsRequest{
		TaskID:        taskID,
		RequestUserID: principal.UserID,
		Limit:         commentsLimit,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"comments": comments.Data}, nil
}

func (h *Handler) handleUpdateTask(ctx context.Context, principal principal, args map[string]any) (map[string]any, error) {
	taskID, err := requiredUUIDArg(args, "task_id")
	if err != nil {
		return nil, err
	}
	projectColumnID, err := requiredUUIDArg(args, "project_column_id")
	if err != nil {
		return nil, err
	}
	title, err := requiredStringArg(args, "title")
	if err != nil {
		return nil, err
	}
	description, err := requiredStringArg(args, "description")
	if err != nil {
		return nil, err
	}
	priority, err := requiredTaskPriorityArg(args, "priority")
	if err != nil {
		return nil, err
	}
	responsibleID, err := optionalUUIDArg(args, "responsible_id")
	if err != nil {
		return nil, err
	}
	dueDate, err := optionalTimeArg(args, "due_date")
	if err != nil {
		return nil, err
	}
	tags, err := optionalStringSliceArg(args, "tags")
	if err != nil {
		return nil, err
	}
	dependsOnTaskIDs, err := optionalUUIDSliceArg(args, "depends_on_task_ids")
	if err != nil {
		return nil, err
	}
	code, err := optionalTrimmedStringArgPointer(args, "code")
	if err != nil {
		return nil, err
	}
	ctx = domain.WithActionOrigin(ctx, domain.ActionOriginMCPAgent)
	task, err := h.taskService.Update(ctx, UpdateTaskRequest{
		TaskID:           taskID,
		Title:            title,
		Description:      description,
		Code:             code,
		ProjectColumnID:  projectColumnID,
		RequestUserID:    principal.UserID,
		Priority:         priority,
		ResponsibleID:    responsibleID,
		DueDate:          dueDate,
		Tags:             tags,
		DependsOnTaskIDs: dependsOnTaskIDs,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"task": task}, nil
}

func (h *Handler) handleMoveTask(ctx context.Context, principal principal, args map[string]any) (map[string]any, error) {
	taskID, err := requiredUUIDArg(args, "task_id")
	if err != nil {
		return nil, err
	}
	projectID, err := requiredUUIDArg(args, "project_id")
	if err != nil {
		return nil, err
	}
	targetColumnID, err := requiredUUIDArg(args, "target_project_column_id")
	if err != nil {
		return nil, err
	}
	afterTaskID, err := optionalUUIDArg(args, "after_task_id")
	if err != nil {
		return nil, err
	}
	ctx = domain.WithActionOrigin(ctx, domain.ActionOriginMCPAgent)
	task, err := h.taskService.Move(ctx, MoveTaskRequest{
		TaskID:          taskID,
		RequestUserID:   principal.UserID,
		AfterTaskID:     afterTaskID,
		ProjectID:       projectID,
		ProjectColumnID: targetColumnID,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"task": task}, nil
}

func (h *Handler) handleAddTaskComment(ctx context.Context, principal principal, args map[string]any) (map[string]any, error) {
	taskID, err := requiredUUIDArg(args, "task_id")
	if err != nil {
		return nil, err
	}
	parentCommentID, err := optionalUUIDArg(args, "parent_comment_id")
	if err != nil {
		return nil, err
	}
	content, err := requiredStringArg(args, "content")
	if err != nil {
		return nil, err
	}
	ctx = domain.WithActionOrigin(ctx, domain.ActionOriginMCPAgent)
	comment, err := h.taskCommentService.Create(ctx, CreateTaskCommentRequest{
		TaskID:          taskID,
		RequestUserID:   principal.UserID,
		Content:         content,
		ParentCommentID: parentCommentID,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"comment": comment}, nil
}

func (h *Handler) handleMarkTaskDone(ctx context.Context, principal principal, args map[string]any) (map[string]any, error) {
	taskID, err := requiredUUIDArg(args, "task_id")
	if err != nil {
		return nil, err
	}
	ctx = domain.WithActionOrigin(ctx, domain.ActionOriginMCPAgent)
	task, err := h.taskService.MarkTaskDone(ctx, MarkTaskDoneRequest{
		TaskID:        taskID,
		RequestUserID: principal.UserID,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"task": task}, nil
}

func (h *Handler) handleAssignTaskToSelf(ctx context.Context, principal principal, args map[string]any) (map[string]any, error) {
	taskID, err := requiredUUIDArg(args, "task_id")
	if err != nil {
		return nil, err
	}
	ctx = domain.WithActionOrigin(ctx, domain.ActionOriginMCPAgent)
	task, err := h.taskService.AssignTaskToSelf(ctx, AssignTaskToSelfRequest{
		TaskID:        taskID,
		RequestUserID: principal.UserID,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"task": task}, nil
}
