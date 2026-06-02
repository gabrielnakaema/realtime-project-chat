package domain

import (
	"time"

	"github.com/google/uuid"
)

type MCPAPIScope string

const (
	MCPAPIScopeProjectsRead      MCPAPIScope = "projects:read"
	MCPAPIScopeProjectsBoardRead MCPAPIScope = "projects:board:read"
	MCPAPIScopeTasksCreate       MCPAPIScope = "tasks:create"
	MCPAPIScopeTasksRead         MCPAPIScope = "tasks:read"
	MCPAPIScopeTasksUpdate       MCPAPIScope = "tasks:update"
	MCPAPIScopeTasksCommentsRead MCPAPIScope = "tasks:comments:read"
	MCPAPIScopeTasksMove         MCPAPIScope = "tasks:move"
	MCPAPIScopeTasksComment      MCPAPIScope = "tasks:comment"
	MCPAPIScopeTasksMarkDone     MCPAPIScope = "tasks:mark_done"
	MCPAPIScopeTasksAssignSelf   MCPAPIScope = "tasks:assign:self"
)

type MCPAPIScopeDefinition struct {
	Scope MCPAPIScope `json:"scope"`
	Label string      `json:"label"`
	Title string      `json:"title"`
}

var MCPAPIScopeDefinitions = []MCPAPIScopeDefinition{
	{
		Scope: MCPAPIScopeProjectsRead,
		Label: "Read projects",
		Title: "View project details and browse the workspace structure.",
	},
	{
		Scope: MCPAPIScopeProjectsBoardRead,
		Label: "Read project boards",
		Title: "Inspect board columns and the tasks grouped inside them.",
	},
	{
		Scope: MCPAPIScopeTasksCreate,
		Label: "Create tasks",
		Title: "Create new tasks in project columns you can access.",
	},
	{
		Scope: MCPAPIScopeTasksRead,
		Label: "Read tasks",
		Title: "Inspect task fields, status, assignments, and related metadata.",
	},
	{
		Scope: MCPAPIScopeTasksUpdate,
		Label: "Update tasks",
		Title: "Edit task details such as title, description, assignee, due date, and tags.",
	},
	{
		Scope: MCPAPIScopeTasksCommentsRead,
		Label: "Read task comments",
		Title: "Load recent task comments when an MCP client requests extra context.",
	},
	{
		Scope: MCPAPIScopeTasksMove,
		Label: "Move tasks",
		Title: "Change where a task sits in the workflow.",
	},
	{
		Scope: MCPAPIScopeTasksComment,
		Label: "Comment on tasks",
		Title: "Post task comments on your behalf from an MCP client.",
	},
	{
		Scope: MCPAPIScopeTasksMarkDone,
		Label: "Mark tasks done",
		Title: "Complete tasks by moving them into a done state when supported.",
	},
	{
		Scope: MCPAPIScopeTasksAssignSelf,
		Label: "Assign tasks to me",
		Title: "Let an MCP client assign tasks to your own account only.",
	},
}

var AllowedMCPAPIScopes = func() []MCPAPIScope {
	allowed := make([]MCPAPIScope, 0, len(MCPAPIScopeDefinitions))
	for _, scope := range MCPAPIScopeDefinitions {
		allowed = append(allowed, scope.Scope)
	}

	return allowed
}()

var allowedMCPAPIScopeSet = func() map[MCPAPIScope]struct{} {
	allowed := make(map[MCPAPIScope]struct{}, len(MCPAPIScopeDefinitions))
	for _, scope := range MCPAPIScopeDefinitions {
		allowed[scope.Scope] = struct{}{}
	}
	return allowed
}()

type MCPAPIKey struct {
	ID         uuid.UUID     `json:"id"`
	UserID     uuid.UUID     `json:"user_id"`
	Name       string        `json:"name"`
	KeyPrefix  string        `json:"key_prefix"`
	SecretHash string        `json:"-"`
	Scopes     []MCPAPIScope `json:"scopes"`
	CreatedAt  time.Time     `json:"created_at"`
	LastUsedAt *time.Time    `json:"last_used_at,omitempty"`
	RevokedAt  *time.Time    `json:"revoked_at,omitempty"`
}

func (k MCPAPIKey) IsRevoked() bool {
	return k.RevokedAt != nil
}

func (k MCPAPIKey) HasScope(scope MCPAPIScope) bool {
	for _, granted := range k.Scopes {
		if granted == scope {
			return true
		}
	}

	return false
}

func (k MCPAPIKey) IsAllowedScope(scope MCPAPIScope) bool {
	_, ok := allowedMCPAPIScopeSet[scope]
	return ok
}
