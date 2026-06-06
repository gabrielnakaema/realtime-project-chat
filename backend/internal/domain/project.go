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
	Columns []ProjectColumn `json:"columns,omitempty"`
}

func (p *Project) IsMember(userId uuid.UUID) bool {
	for _, member := range p.Members {
		if member.UserId == userId {
			return true
		}
	}
	return false
}

func (p *Project) IsCreator(userId uuid.UUID) bool {
	for _, member := range p.Members {
		if member.UserId == userId && member.Role == ProjectMemberRoleCreator {
			return true
		}
	}
	return false
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

type ProjectColumn struct {
	Id           uuid.UUID `json:"id"`
	ProjectId    uuid.UUID `json:"project_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Color        string    `json:"color"`
	Position     int       `json:"position"`
	IsDoneColumn bool      `json:"is_done_column"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
