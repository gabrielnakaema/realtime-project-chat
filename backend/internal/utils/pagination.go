package utils

import (
	"time"

	"github.com/google/uuid"
)

type CursorPaginated[T any] struct {
	Data    []T  `json:"data"`
	HasNext bool `json:"has_next"`
}

type PaginationBeforeParams struct {
	Before time.Time `json:"before"`
	Id     uuid.UUID `json:"id"`
	Limit  int32     `json:"limit"`
}
