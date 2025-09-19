package utils

import "time"

type CursorPaginated[T any] struct {
	Data    []T  `json:"data"`
	HasNext bool `json:"has_next"`
}

type PaginationBeforeParams struct {
	Before *time.Time `json:"before"`
	Limit  int32      `json:"limit"`
}
