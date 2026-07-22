package mcp

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/google/uuid"
)

const maxToolResultLimit = 100

func paramsFromRaw(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}

	var params map[string]any
	if err := json.Unmarshal(raw, &params); err != nil {
		return map[string]any{}
	}

	return params
}

func requiredUUIDArg(args map[string]any, key string) (uuid.UUID, error) {
	raw, err := requiredStringArg(args, key)
	if err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be a valid uuid", key))
	}

	return id, nil
}

func optionalUUIDArg(args map[string]any, key string) (*uuid.UUID, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}

	value, ok := raw.(string)
	if !ok || strings.TrimSpace(value) == "" {
		return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be a valid uuid", key))
	}

	id, err := uuid.Parse(value)
	if err != nil {
		return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be a valid uuid", key))
	}

	return &id, nil
}

func optionalUUIDSliceArg(args map[string]any, key string) ([]uuid.UUID, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}

	items, ok := raw.([]any)
	if !ok {
		return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be an array of uuids", key))
	}

	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be an array of uuids", key))
		}
		id, err := uuid.Parse(value)
		if err != nil {
			return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be an array of uuids", key))
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func requiredStringArg(args map[string]any, key string) (string, error) {
	raw, ok := args[key]
	if !ok {
		return "", apperr.BusinessValidationError(fmt.Sprintf("%s is required", key))
	}

	value, ok := raw.(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", apperr.BusinessValidationError(fmt.Sprintf("%s is required", key))
	}

	return strings.TrimSpace(value), nil
}

func optionalLimitArg(args map[string]any, key string, defaultValue int) (int, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return defaultValue, nil
	}

	number, ok := raw.(float64)
	if !ok || number < 1 || number > maxToolResultLimit || math.Trunc(number) != number {
		return 0, apperr.BusinessValidationError(fmt.Sprintf("%s must be an integer between 1 and %d", key, maxToolResultLimit))
	}

	return int(number), nil
}

func optionalBoolArg(args map[string]any, key string, defaultValue bool) (bool, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return defaultValue, nil
	}

	value, ok := raw.(bool)
	if !ok {
		return false, apperr.BusinessValidationError(fmt.Sprintf("%s must be a boolean", key))
	}

	return value, nil
}

func requiredTaskPriorityArg(args map[string]any, key string) (domain.TaskPriority, error) {
	value, err := requiredStringArg(args, key)
	if err != nil {
		return "", err
	}

	priority := domain.TaskPriority(strings.ToLower(value))
	for _, allowed := range domain.AllowedTaskPriorities {
		if priority == allowed {
			return priority, nil
		}
	}

	return "", apperr.BusinessValidationError(fmt.Sprintf("%s must be one of: low, medium, high", key))
}

func optionalTimeArg(args map[string]any, key string) (*time.Time, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}

	value, ok := raw.(string)
	if !ok || strings.TrimSpace(value) == "" {
		return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be a valid RFC3339 datetime", key))
	}

	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be a valid RFC3339 datetime", key))
	}

	return &parsed, nil
}

func optionalStringSliceArg(args map[string]any, key string) ([]string, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}

	items, ok := raw.([]any)
	if !ok {
		return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be an array of strings", key))
	}

	values := make([]string, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be an array of strings", key))
		}
		values = append(values, strings.TrimSpace(value))
	}

	return values, nil
}

func optionalTrimmedStringArg(args map[string]any, key string) (string, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return "", nil
	}

	value, ok := raw.(string)
	if !ok {
		return "", apperr.BusinessValidationError(fmt.Sprintf("%s must be a string", key))
	}

	return strings.TrimSpace(value), nil
}

func optionalTrimmedStringArgPointer(args map[string]any, key string) (*string, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}

	value, ok := raw.(string)
	if !ok {
		return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be a string", key))
	}

	trimmed := strings.TrimSpace(value)
	return &trimmed, nil
}
