package domain

import "context"

type ActionOrigin string

const (
	ActionOriginUser     ActionOrigin = "user"
	ActionOriginMCPAgent ActionOrigin = "mcp_agent"
)

func (origin ActionOrigin) OrUser() ActionOrigin {
	if origin == "" {
		return ActionOriginUser
	}

	return origin
}

type actionOriginContextKey string

const ActionOriginContextKey actionOriginContextKey = "action_origin"

func WithActionOrigin(ctx context.Context, origin ActionOrigin) context.Context {
	return context.WithValue(ctx, ActionOriginContextKey, origin.OrUser())
}

func ActionOriginFromContext(ctx context.Context) ActionOrigin {
	origin, ok := ctx.Value(ActionOriginContextKey).(ActionOrigin)
	if !ok {
		return ActionOriginUser
	}

	return origin.OrUser()
}
