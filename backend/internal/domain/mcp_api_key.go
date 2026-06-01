package domain

import (
	"time"

	"github.com/google/uuid"
)

type MCPAPIScope string

const (
	MCPAPIScopeProjectsRead    MCPAPIScope = "projects:read"
	MCPAPIScopeTasksRead       MCPAPIScope = "tasks:read"
	MCPAPIScopeTasksMove       MCPAPIScope = "tasks:move"
	MCPAPIScopeTasksComment    MCPAPIScope = "tasks:comment"
	MCPAPIScopeTasksMarkDone   MCPAPIScope = "tasks:mark_done"
	MCPAPIScopeTasksAssignSelf MCPAPIScope = "tasks:assign:self"
)

var AllowedMCPAPIScopes = []MCPAPIScope{
	MCPAPIScopeProjectsRead,
	MCPAPIScopeTasksRead,
	MCPAPIScopeTasksMove,
	MCPAPIScopeTasksComment,
	MCPAPIScopeTasksMarkDone,
	MCPAPIScopeTasksAssignSelf,
}

var allowedMCPAPIScopeSet = func() map[MCPAPIScope]struct{} {
	allowed := make(map[MCPAPIScope]struct{}, len(AllowedMCPAPIScopes))
	for _, scope := range AllowedMCPAPIScopes {
		allowed[scope] = struct{}{}
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
