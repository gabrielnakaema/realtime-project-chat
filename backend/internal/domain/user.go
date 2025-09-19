package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	Email     string    `json:"email,omitempty"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type RefreshToken struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	Active    bool
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}
