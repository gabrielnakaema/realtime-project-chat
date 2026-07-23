package mcp

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/google/uuid"
)

type searchTasksCursor struct {
	DueDate   *time.Time `json:"due_date"`
	UpdatedAt time.Time  `json:"updated_at"`
	TaskID    uuid.UUID  `json:"task_id"`
}

func encodeSearchTasksCursor(cursor searchTasksCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeSearchTasksCursor(value string) (*searchTasksCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, apperr.BusinessValidationError("cursor is invalid")
	}

	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	var cursor searchTasksCursor
	if err := decoder.Decode(&cursor); err != nil || cursor.UpdatedAt.IsZero() || cursor.TaskID == uuid.Nil {
		return nil, apperr.BusinessValidationError("cursor is invalid")
	}
	return &cursor, nil
}
