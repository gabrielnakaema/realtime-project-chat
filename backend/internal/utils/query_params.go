package utils

import (
	"net/http"
	"strconv"
)

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
