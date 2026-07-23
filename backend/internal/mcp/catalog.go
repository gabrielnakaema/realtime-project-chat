package mcp

import (
	"github.com/gabrielnakaema/project-chat/internal/domain"
)

type toolSpec struct {
	Name          string
	Title         string
	Description   string
	RequiredScope domain.MCPAPIScope
	InputSchema   map[string]any
	OutputSchema  map[string]any
	Handle        toolHandlerFunc
	SuccessText   toolSuccessTextFunc
}

var registeredTools = buildToolCatalog()

func toolDefinitionsForPrincipal(principal principal) []map[string]any {
	tools := make([]map[string]any, 0, len(registeredTools))
	for _, spec := range registeredTools {
		if !principal.HasScope(spec.RequiredScope) {
			continue
		}

		tools = append(tools, map[string]any{
			"name":         spec.Name,
			"title":        spec.Title,
			"description":  spec.Description,
			"inputSchema":  spec.InputSchema,
			"outputSchema": spec.OutputSchema,
		})
	}

	return tools
}

func findToolSpec(name string) (toolSpec, bool) {
	for _, spec := range registeredTools {
		if spec.Name == name {
			return spec, true
		}
	}

	return toolSpec{}, false
}

func buildToolCatalog() []toolSpec {
	taskSchema := map[string]any{
		"type":        "object",
		"description": "Task record returned by Project Chat, including identifiers, status, column placement, priority, tags, timestamps, and related metadata.",
	}
	projectSchema := map[string]any{
		"type":        "object",
		"description": "Project record returned by Project Chat, including columns, members, and top-level metadata.",
	}
	commentSchema := map[string]any{
		"type":        "object",
		"description": "Task comment record returned by Project Chat, including content, timestamps, author, and reply metadata.",
	}

	return []toolSpec{
		{
			Name:          "list_projects",
			Title:         "List Visible Projects",
			Description:   "List the projects visible to the authenticated user so the client can discover project ids before reading boards or tasks.",
			RequiredScope: domain.MCPAPIScopeProjectsRead,
			Handle:        (*Handler).handleListProjects,
			SuccessText:   listProjectsSuccessText,
			InputSchema: map[string]any{
				"type":                 "object",
				"description":          "No arguments are required.",
				"properties":           map[string]any{},
				"additionalProperties": false,
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"projects": map[string]any{
						"type":        "array",
						"description": "Projects visible to this API key.",
						"items":       projectSchema,
					},
				},
				"required": []string{"projects"},
			},
		},
		{
			Name:          "list_project_board",
			Title:         "List Board Columns And Tasks",
			Description:   "Load a project's board structure and grouped tasks. Use this after list_projects when you need the live workflow state.",
			RequiredScope: domain.MCPAPIScopeProjectsBoardRead,
			Handle:        (*Handler).handleListProjectBoard,
			SuccessText:   projectBoardSuccessText,
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"project_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "The project id returned by list_projects.",
					},
					"project_column_ids": map[string]any{
						"type":        "array",
						"description": "Optional subset of column ids to load instead of the full board.",
						"items": map[string]any{
							"type":   "string",
							"format": "uuid",
						},
					},
					"limit_per_column": map[string]any{
						"type":        "integer",
						"minimum":     1,
						"maximum":     maxToolResultLimit,
						"description": "Maximum number of tasks to return per column. Defaults to 15.",
					},
					"include_archived": map[string]any{
						"type":        "boolean",
						"description": "Whether archived tasks should be included. Defaults to false.",
					},
				},
				"required":             []string{"project_id"},
				"additionalProperties": false,
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"project": projectSchema,
					"columns": map[string]any{
						"type":        "array",
						"description": "Ordered project columns for the board.",
						"items": map[string]any{
							"type":        "object",
							"description": "Project column metadata including id, name, description, color, and done-column flag.",
						},
					},
					"tasks_by_column": map[string]any{
						"type":        "object",
						"description": "Paginated task groups keyed by project_column_id.",
					},
				},
				"required": []string{"project", "columns", "tasks_by_column"},
			},
		},
		{
			Name:          "search_tasks",
			Title:         "Search Tasks In Project Columns",
			Description:   "Search for case-insensitive partial matches in task titles, descriptions, and codes within selected project columns. Discover project and column ids with list_project_board first.",
			RequiredScope: domain.MCPAPIScopeTasksRead,
			Handle:        (*Handler).handleSearchTasks,
			SuccessText:   searchTasksSuccessText,
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"project_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Project id that owns every selected column.",
					},
					"project_column_ids": map[string]any{
						"type":        "array",
						"minItems":    1,
						"uniqueItems": true,
						"description": "One or more project column ids in which to search. Done columns are searchable when selected.",
						"items": map[string]any{
							"type":   "string",
							"format": "uuid",
						},
					},
					"query": map[string]any{
						"type":        "string",
						"minLength":   1,
						"description": "Text matched as a case-insensitive substring against task title, description, or code.",
					},
					"include_archived": map[string]any{
						"type":        "boolean",
						"description": "Include archived matches in addition to active matches. Only the project creator may enable this. Defaults to false.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"minimum":     1,
						"maximum":     maxToolResultLimit,
						"description": "Maximum number of matching tasks to return. Defaults to 25.",
					},
					"cursor": map[string]any{
						"type":        "string",
						"minLength":   1,
						"description": "Opaque continuation cursor returned by a previous search_tasks call with the same filters.",
					},
				},
				"required":             []string{"project_id", "project_column_ids", "query"},
				"additionalProperties": false,
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"tasks": map[string]any{
						"type":        "array",
						"description": "Matching tasks in stable due-date order.",
						"items":       taskSchema,
					},
					"has_next": map[string]any{
						"type":        "boolean",
						"description": "Whether another result page is available.",
					},
					"next_cursor": map[string]any{
						"description": "Opaque continuation cursor, or null on the final page.",
						"anyOf": []map[string]any{
							{"type": "string"},
							{"type": "null"},
						},
					},
				},
				"required": []string{"tasks", "has_next", "next_cursor"},
			},
		},
		{
			Name:          "create_task",
			Title:         "Create Task",
			Description:   "Create a new task in a specific project column.",
			RequiredScope: domain.MCPAPIScopeTasksCreate,
			Handle:        (*Handler).handleCreateTask,
			SuccessText:   taskSuccessText("Created task %q."),
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"project_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Project id that will own the task.",
					},
					"project_column_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Column id where the new task should start.",
					},
					"title": map[string]any{
						"type":        "string",
						"description": "Short task title.",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "Rich-text task description stored by Project Chat.",
					},
					"code": map[string]any{
						"type":        "string",
						"description": "Optional short task code such as BACKEND-3.",
					},
					"priority": map[string]any{
						"type":        "string",
						"enum":        []string{"low", "medium", "high"},
						"description": "Task priority.",
					},
					"responsible_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Optional assignee user id.",
					},
					"due_date": map[string]any{
						"type":        "string",
						"format":      "date-time",
						"description": "Optional RFC3339 due date.",
					},
					"tags": map[string]any{
						"type":        "array",
						"description": "Optional task tags.",
						"items": map[string]any{
							"type": "string",
						},
					},
					"depends_on_task_ids": map[string]any{
						"type":        "array",
						"description": "Optional task ids that must be completed before this task.",
						"items": map[string]any{
							"type":   "string",
							"format": "uuid",
						},
					},
				},
				"required":             []string{"project_id", "project_column_id", "title", "description", "priority"},
				"additionalProperties": false,
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task": taskSchema,
				},
				"required": []string{"task"},
			},
		},
		{
			Name:          "find_task_by_code",
			Title:         "Find Task By Code",
			Description:   "Find a single active task in a project by its exact code. If the same code is used by multiple active tasks in that project, the tool returns an ambiguity error instead of guessing.",
			RequiredScope: domain.MCPAPIScopeTasksRead,
			Handle:        (*Handler).handleFindTaskByCode,
			SuccessText:   taskSuccessText("Found task %q by code."),
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"project_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Project id that owns the task.",
					},
					"code": map[string]any{
						"type":        "string",
						"description": "Exact task code to match within the project, for example BACKEND-5.",
					},
					"include_comments": map[string]any{
						"type":        "boolean",
						"description": "When true, recent comments are attached to the task result and require the tasks:comments:read scope.",
					},
					"comments_limit": map[string]any{
						"type":        "integer",
						"minimum":     1,
						"maximum":     maxToolResultLimit,
						"description": "Maximum number of comments to attach when include_comments is true. Defaults to 10.",
					},
				},
				"required":             []string{"project_id", "code"},
				"additionalProperties": false,
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task": taskSchema,
				},
				"required": []string{"task"},
			},
		},
		{
			Name:          "get_task",
			Title:         "Get Task",
			Description:   "Load one task by id. When include_comments is true, the API key must also have the tasks:comments:read scope.",
			RequiredScope: domain.MCPAPIScopeTasksRead,
			Handle:        (*Handler).handleGetTask,
			SuccessText:   taskSuccessText("Loaded task %q."),
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Task id to load.",
					},
					"include_comments": map[string]any{
						"type":        "boolean",
						"description": "When true, recent comments are attached to the task result and require the tasks:comments:read scope.",
					},
					"comments_limit": map[string]any{
						"type":        "integer",
						"minimum":     1,
						"maximum":     maxToolResultLimit,
						"description": "Maximum number of comments to attach when include_comments is true. Defaults to 10.",
					},
				},
				"required":             []string{"task_id"},
				"additionalProperties": false,
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task": taskSchema,
				},
				"required": []string{"task"},
			},
		},
		{
			Name:          "list_task_comments",
			Title:         "List Task Comments",
			Description:   "List recent comments for a task without loading the full task payload.",
			RequiredScope: domain.MCPAPIScopeTasksCommentsRead,
			Handle:        (*Handler).handleListTaskComments,
			SuccessText:   listTaskCommentsSuccessText,
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Task id whose comments should be listed.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"minimum":     1,
						"maximum":     maxToolResultLimit,
						"description": "Maximum number of comments to return. Defaults to 10.",
					},
				},
				"required":             []string{"task_id"},
				"additionalProperties": false,
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"comments": map[string]any{
						"type":        "array",
						"description": "Recent task comments ordered by the service pagination rules.",
						"items":       commentSchema,
					},
				},
				"required": []string{"comments"},
			},
		},
		{
			Name:          "update_task",
			Title:         "Update Task",
			Description:   "Update an existing task's editable fields while keeping it in a specified project column.",
			RequiredScope: domain.MCPAPIScopeTasksUpdate,
			Handle:        (*Handler).handleUpdateTask,
			SuccessText:   taskSuccessText("Updated task %q."),
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Task id to update.",
					},
					"project_column_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Column id the task should belong to after the update.",
					},
					"title": map[string]any{
						"type":        "string",
						"description": "Updated task title.",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "Updated rich-text task description.",
					},
					"code": map[string]any{
						"type":        "string",
						"description": "Optional updated task code.",
					},
					"priority": map[string]any{
						"type":        "string",
						"enum":        []string{"low", "medium", "high"},
						"description": "Updated task priority.",
					},
					"responsible_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Optional updated assignee user id.",
					},
					"due_date": map[string]any{
						"type":        "string",
						"format":      "date-time",
						"description": "Optional updated RFC3339 due date.",
					},
					"tags": map[string]any{
						"type":        "array",
						"description": "Updated task tags.",
						"items": map[string]any{
							"type": "string",
						},
					},
					"depends_on_task_ids": map[string]any{
						"type":        "array",
						"description": "Updated task ids that must be completed before this task.",
						"items": map[string]any{
							"type":   "string",
							"format": "uuid",
						},
					},
				},
				"required":             []string{"task_id", "project_column_id", "title", "description", "priority"},
				"additionalProperties": false,
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task": taskSchema,
				},
				"required": []string{"task"},
			},
		},
		{
			Name:          "move_task",
			Title:         "Move Task",
			Description:   "Move a task into another project column, optionally after another task in that destination column.",
			RequiredScope: domain.MCPAPIScopeTasksMove,
			Handle:        (*Handler).handleMoveTask,
			SuccessText:   taskSuccessText("Moved task %q to a new column."),
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Task id to move.",
					},
					"project_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Project id that owns the task and destination column.",
					},
					"target_project_column_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Destination column id.",
					},
					"after_task_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Optional task id to position the moved task after within the destination column.",
					},
				},
				"required":             []string{"task_id", "project_id", "target_project_column_id"},
				"additionalProperties": false,
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task": taskSchema,
				},
				"required": []string{"task"},
			},
		},
		{
			Name:          "add_task_comment",
			Title:         "Add Task Comment",
			Description:   "Add a new task comment, optionally as a reply to an existing comment.",
			RequiredScope: domain.MCPAPIScopeTasksComment,
			Handle:        (*Handler).handleAddTaskComment,
			SuccessText:   staticSuccessText("Added a task comment."),
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Task id that will receive the comment.",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "Comment body.",
					},
					"parent_comment_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Optional parent comment id when posting a reply.",
					},
				},
				"required":             []string{"task_id", "content"},
				"additionalProperties": false,
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"comment": commentSchema,
				},
				"required": []string{"comment"},
			},
		},
		{
			Name:          "mark_task_done",
			Title:         "Mark Task Done",
			Description:   "Move a task into the project's configured done column.",
			RequiredScope: domain.MCPAPIScopeTasksMarkDone,
			Handle:        (*Handler).handleMarkTaskDone,
			SuccessText:   taskSuccessText("Marked task %q done."),
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Task id to complete.",
					},
				},
				"required":             []string{"task_id"},
				"additionalProperties": false,
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task": taskSchema,
				},
				"required": []string{"task"},
			},
		},
		{
			Name:          "assign_task_to_self",
			Title:         "Assign Task To Authenticated User",
			Description:   "Assign the task to the same user who owns the MCP API key. This tool never assigns other users.",
			RequiredScope: domain.MCPAPIScopeTasksAssignSelf,
			Handle:        (*Handler).handleAssignTaskToSelf,
			SuccessText:   taskSuccessText("Assigned task %q to the authenticated user."),
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task_id": map[string]any{
						"type":        "string",
						"format":      "uuid",
						"description": "Task id to self-assign.",
					},
				},
				"required":             []string{"task_id"},
				"additionalProperties": false,
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task": taskSchema,
				},
				"required": []string{"task"},
			},
		},
	}
}
