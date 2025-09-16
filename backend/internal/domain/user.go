package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type RefreshToken struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	Active    bool
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}
