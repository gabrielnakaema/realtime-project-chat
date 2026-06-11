package mcp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gabrielnakaema/project-chat/internal/domain"
)

const (
	serverGuideURI = "project-chat://server/guide"
	scopeGuideURI  = "project-chat://server/scopes"
)

type toolSpec struct {
	Name          string
	Title         string
	Description   string
	RequiredScope domain.MCPAPIScope
	InputSchema   map[string]any
	OutputSchema  map[string]any
}

func initializeInstructions(principal principal) string {
	return fmt.Sprintf(
		"Project Chat MCP exposes boards, tasks, and task comments. Start with list_projects, then list_project_board or get_task. Read %s for workflow guidance and %s for granted scope details. tools/list only returns tools allowed by this API key.",
		serverGuideURI,
		scopeGuideURI,
	)
}

func toolDefinitionsForPrincipal(principal principal) []map[string]any {
	catalog := toolCatalog()
	tools := make([]map[string]any, 0, len(catalog))
	for _, spec := range catalog {
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

func resourceDefinitionsForPrincipal(principal principal) []map[string]any {
	return []map[string]any{
		{
			"uri":         serverGuideURI,
			"name":        "Project Chat MCP Guide",
			"description": "Workflow guidance for discovering projects, boards, tasks, and comments through this MCP server.",
			"mimeType":    "text/markdown",
		},
		{
			"uri":         scopeGuideURI,
			"name":        "Project Chat MCP Scope Guide",
			"description": "The scopes granted to this API key and what each scope allows the client to do.",
			"mimeType":    "text/markdown",
		},
	}
}

func readResource(uri string, principal principal) (map[string]any, error) {
	var text string

	switch uri {
	case serverGuideURI:
		text = buildServerGuide(principal)
	case scopeGuideURI:
		text = buildScopeGuide(principal)
	default:
		return nil, domain.NotFoundError("resource not found")
	}

	return map[string]any{
		"contents": []map[string]any{
			{
				"uri":      uri,
				"mimeType": "text/markdown",
				"text":     text,
			},
		},
	}, nil
}

func buildServerGuide(principal principal) string {
	var builder strings.Builder

	builder.WriteString("# Project Chat MCP Guide\n\n")
	builder.WriteString("This server exposes Project Chat workspace data over MCP. It is designed for board and task workflows rather than general code browsing.\n\n")
	builder.WriteString("## Recommended workflow\n")
	builder.WriteString("1. Call `list_projects` to discover which projects this API key can access.\n")
	builder.WriteString("2. Call `find_task_by_code` when you know a project-scoped task code and need the matching task id quickly.\n")
	builder.WriteString("3. Call `list_project_board` to inspect project columns and the tasks grouped inside them.\n")
	builder.WriteString("4. Call `get_task` when you need the full task record. Set `include_comments=true` only when you also have the `tasks:comments:read` scope.\n")
	builder.WriteString("5. Use any write action returned by `tools/list` only when the capability is visible for this API key.\n\n")
	builder.WriteString("## Important notes\n")
	builder.WriteString("- `tools/list` is scope-aware. If a capability is missing from the list, this API key cannot use it.\n")
	builder.WriteString("- `list_project_board` is the best way to understand the board layout without scraping the web UI or repository.\n")
	builder.WriteString("- `mark_task_done` moves the task into the project's configured done column.\n")
	builder.WriteString("- `assign_task_to_self` only assigns the authenticated user.\n\n")
	builder.WriteString("## Available tools for this API key\n")
	for _, spec := range toolCatalog() {
		if !principal.HasScope(spec.RequiredScope) {
			continue
		}
		builder.WriteString(fmt.Sprintf("- `%s`: %s\n", spec.Name, spec.Description))
	}

	return builder.String()
}

func buildScopeGuide(principal principal) string {
	var builder strings.Builder

	builder.WriteString("# Project Chat MCP Scope Guide\n\n")
	builder.WriteString("The following scopes are granted to this API key. Missing scopes are intentionally omitted from `tools/list`.\n\n")

	for _, definition := range domain.MCPAPIScopeDefinitions {
		if !principal.HasScope(definition.Scope) {
			continue
		}
		builder.WriteString(fmt.Sprintf("- `%s` (%s): %s\n", definition.Scope, definition.Label, definition.Title))
	}

	return builder.String()
}

func toolSuccessText(name string, result map[string]any) string {
	switch name {
	case "list_projects":
		if projects, ok := result["projects"].([]domain.Project); ok {
			return fmt.Sprintf("Listed %d visible project(s).", len(projects))
		}
	case "list_project_board":
		if project, ok := result["project"].(*domain.Project); ok {
			return fmt.Sprintf("Loaded board for project %q with %d column(s).", project.Name, len(project.Columns))
		}
	case "create_task":
		if task, ok := result["task"].(*domain.Task); ok {
			return fmt.Sprintf("Created task %q.", task.Title)
		}
	case "get_task":
		if task, ok := result["task"].(*domain.Task); ok {
			return fmt.Sprintf("Loaded task %q.", task.Title)
		}
	case "find_task_by_code":
		if task, ok := result["task"].(*domain.Task); ok {
			return fmt.Sprintf("Found task %q by code.", task.Title)
		}
	case "list_task_comments":
		if comments, ok := result["comments"].([]domain.TaskComment); ok {
			return fmt.Sprintf("Listed %d task comment(s).", len(comments))
		}
	case "update_task":
		if task, ok := result["task"].(*domain.Task); ok {
			return fmt.Sprintf("Updated task %q.", task.Title)
		}
	case "move_task":
		if task, ok := result["task"].(*domain.Task); ok {
			return fmt.Sprintf("Moved task %q to a new column.", task.Title)
		}
	case "add_task_comment":
		return "Added a task comment."
	case "mark_task_done":
		if task, ok := result["task"].(*domain.Task); ok {
			return fmt.Sprintf("Marked task %q done.", task.Title)
		}
	case "assign_task_to_self":
		if task, ok := result["task"].(*domain.Task); ok {
			return fmt.Sprintf("Assigned task %q to the authenticated user.", task.Title)
		}
	}

	return name + " completed successfully"
}

func toolResultContent(summary string, structured any) ([]map[string]any, error) {
	jsonBytes, err := json.Marshal(structured)
	if err != nil {
		return nil, fmt.Errorf("marshal tool structured content: %w", err)
	}

	return []map[string]any{
		{
			"type": "text",
			"text": summary,
		},
		{
			"type": "text",
			"text": string(jsonBytes),
		},
	}, nil
}

func toolSuccessResult(name string, result map[string]any) (map[string]any, error) {
	content, err := toolResultContent(toolSuccessText(name, result), result)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"content":           content,
		"structuredContent": result,
	}, nil
}

func toolCatalog() []toolSpec {
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
						"maximum":     100,
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
			Name:          "create_task",
			Title:         "Create Task",
			Description:   "Create a new task in a specific project column.",
			RequiredScope: domain.MCPAPIScopeTasksCreate,
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
						"maximum":     100,
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
						"maximum":     100,
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
						"maximum":     100,
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
