package mcp

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
)

func toolErrorResult(err error) map[string]any {
	var domainErr apperr.DomainError
	if !errors.As(err, &domainErr) {
		return buildToolErrorResult("internal server error", "server_error", "internal server error")
	}

	errorType := "server_error"
	switch domainErr.Code {
	case apperr.UnauthorizedErrorCode:
		errorType = "authentication_failed"
	case apperr.ForbiddenErrorCode:
		if domainErr.Message == "missing required scope" {
			errorType = "missing_scope"
		} else {
			errorType = "forbidden"
		}
	case apperr.NotFoundErrorCode:
		errorType = "not_found"
	case apperr.BusinessValidationErrorCode:
		errorType = "business_validation"
	}

	return buildToolErrorResult(domainErr.Message, errorType, domainErr.Message)
}

func buildToolErrorResult(summary string, errorType string, message string) map[string]any {
	structuredContent := map[string]any{
		"error": map[string]any{
			"type":    errorType,
			"message": message,
		},
	}

	content, err := toolResultContent(summary, structuredContent)
	if err != nil {
		return map[string]any{
			"content": []map[string]any{{"type": "text", "text": summary}},
			"isError": true,
			"structuredContent": map[string]any{
				"error": map[string]any{
					"type":    "server_error",
					"message": summary,
				},
			},
		}
	}

	return map[string]any{
		"content":           content,
		"isError":           true,
		"structuredContent": structuredContent,
	}
}

type toolSuccessTextFunc func(result map[string]any) string

func toolSuccessText(name string, result map[string]any) string {
	if spec, ok := findToolSpec(name); ok && spec.SuccessText != nil {
		if summary := spec.SuccessText(result); summary != "" {
			return summary
		}
	}

	return name + " completed successfully"
}

func listProjectsSuccessText(result map[string]any) string {
	projects, ok := result["projects"].([]domain.Project)
	if !ok {
		return ""
	}
	return fmt.Sprintf("Listed %d visible project(s).", len(projects))
}

func projectBoardSuccessText(result map[string]any) string {
	project, ok := result["project"].(*domain.Project)
	if !ok {
		return ""
	}
	return fmt.Sprintf("Loaded board for project %q with %d column(s).", project.Name, len(project.Columns))
}

func searchTasksSuccessText(result map[string]any) string {
	tasks, ok := result["tasks"].([]domain.Task)
	if !ok {
		return ""
	}
	return fmt.Sprintf("Found %d matching task(s).", len(tasks))
}

func listTaskCommentsSuccessText(result map[string]any) string {
	comments, ok := result["comments"].([]domain.TaskComment)
	if !ok {
		return ""
	}
	return fmt.Sprintf("Listed %d task comment(s).", len(comments))
}

func taskSuccessText(format string) toolSuccessTextFunc {
	return func(result map[string]any) string {
		task, ok := result["task"].(*domain.Task)
		if !ok {
			return ""
		}
		return fmt.Sprintf(format, task.Title)
	}
}

func staticSuccessText(summary string) toolSuccessTextFunc {
	return func(map[string]any) string {
		return summary
	}
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
