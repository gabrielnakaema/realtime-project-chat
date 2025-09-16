package domain

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	Id          uuid.UUID `json:"id"`
	UserId      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Members []ProjectMember `json:"members,omitempty"`
}

type ProjectMemberRole string

var (
	ProjectMemberRoleCreator ProjectMemberRole = "creator"
	ProjectMemberRoleMember  ProjectMemberRole = "member"
)

type ProjectMember struct {
	Id        uuid.UUID         `json:"id"`
	UserId    uuid.UUID         `json:"user_id"`
	User      *User             `json:"user,omitempty"`
	ProjectId uuid.UUID         `json:"project_id"`
	Role      ProjectMemberRole `json:"role"`
}
