package mcp

import (
	"fmt"
	"strings"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
)

const (
	serverGuideURI = "project-chat://server/guide"
	scopeGuideURI  = "project-chat://server/scopes"
)

func initializeInstructions(principal principal) string {
	return fmt.Sprintf(
		"Project Chat MCP exposes boards, tasks, and task comments. Start with list_projects, then list_project_board before searching selected columns with search_tasks. Read %s for workflow guidance and %s for granted scope details. tools/list only returns tools allowed by this API key.",
		serverGuideURI,
		scopeGuideURI,
	)
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
		return nil, apperr.NotFoundError("resource not found")
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
	builder.WriteString("2. Call `list_project_board` to inspect project columns and discover the column ids used by task search.\n")
	builder.WriteString("3. Call `search_tasks` with a project id, one or more column ids, and text when you need partial matches in task titles, descriptions, or codes.\n")
	builder.WriteString("4. Call `find_task_by_code` when you know a project-scoped task code and need the exact matching task id quickly.\n")
	builder.WriteString("5. Call `get_task` when you need the full task record. Set `include_comments=true` only when you also have the `tasks:comments:read` scope.\n")
	builder.WriteString("6. Use any write action returned by `tools/list` only when the capability is visible for this API key.\n\n")
	builder.WriteString("## Important notes\n")
	builder.WriteString("- `tools/list` is scope-aware. If a capability is missing from the list, this API key cannot use it.\n")
	builder.WriteString("- `list_project_board` is the best way to understand the board layout without scraping the web UI or repository.\n")
	builder.WriteString("- `search_tasks` requires column ids from the selected project and returns an opaque cursor when more matches are available.\n")
	builder.WriteString("- `mark_task_done` moves the task into the project's configured done column.\n")
	builder.WriteString("- `assign_task_to_self` only assigns the authenticated user.\n\n")
	builder.WriteString("## Available tools for this API key\n")
	for _, spec := range registeredTools {
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
