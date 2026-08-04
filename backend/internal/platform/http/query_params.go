package platformhttp

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ParseURLUUID(r *http.Request, key string) (uuid.UUID, error) {
	value := chi.URLParam(r, key)
	if value == "" {
		return uuid.Nil, fmt.Errorf("%s is required", key)
	}

	return uuid.Parse(value)
}

func ParseOptionalQueryUUID(r *http.Request, key string) (*uuid.UUID, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil, nil
	}

	parsed, err := uuid.Parse(value)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func ParseLimit(r *http.Request, defaultValue, maxValue int32) (int32, error) {
	value := r.URL.Query().Get("limit")
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, errors.New("limit must be a valid integer")
	}

	limit := int32(parsed)
	if limit <= 0 {
		return 0, errors.New("limit must be greater than 0")
	}

	if limit > maxValue {
		return 0, fmt.Errorf("limit must be less than or equal to %d", maxValue)
	}

	return limit, nil
}

func ParseRFC3339Cursor(r *http.Request, key string) (*time.Time, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func GetQueryString(r *http.Request, key string, defaultValue string) string {
	query := r.URL.Query()
	value := query.Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetQueryInt(r *http.Request, key string, defaultValue int32) int32 {
	query := r.URL.Query()
	value := query.Get(key)
	if value == "" {
		return defaultValue
	}

	parsedValue, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return defaultValue
	}

	return int32(parsedValue)
}

func GetQueryFloat(r *http.Request, key string, defaultValue float64) float64 {
	query := r.URL.Query()
	value := query.Get(key)
	if value == "" {
		return defaultValue
	}
	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return parsedValue
}
