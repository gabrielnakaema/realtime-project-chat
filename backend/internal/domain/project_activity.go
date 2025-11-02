package domain

import (
	"time"

	"github.com/google/uuid"
)

type ActivityType string

var (
	TaskCreated          ActivityType = "task.created"
	TaskUpdated          ActivityType = "task.updated"
	TaskDeleted          ActivityType = "task.deleted"
	ProjectCreated       ActivityType = "project.created"
	ProjectUpdated       ActivityType = "project.updated"
	ProjectMemberCreated ActivityType = "project.member.created"
	ProjectMemberDeleted ActivityType = "project.member.deleted"
)

type EntityType string

var (
	TaskEntityType          EntityType = "task"
	ProjectEntityType       EntityType = "project"
	ProjectMemberEntityType EntityType = "project.member"
)

type ProjectActivity struct {
	Id            uuid.UUID      `json:"id"`
	Project       Project        `json:"project"`
	Actor         User           `json:"actor"`
	ActivityType  ActivityType   `json:"activity_type"`
	ActivityData  any            `json:"activity_data"`
	EntityType    EntityType     `json:"entity_type"`
	EntityId      uuid.UUID      `json:"entity_id"`
	Task          *Task          `json:"task,omitempty"`
	ProjectMember *ProjectMember `json:"project_member,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func ProjectCreatedActivity(project Project, actor User) ProjectActivity {
	return ProjectActivity{
		Project:      project,
		ActivityType: ProjectCreated,
		Actor:        actor,
		ActivityData: make(map[string]interface{}),
		EntityType:   ProjectEntityType,
		EntityId:     project.Id,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TaskCreatedActivity(project Project, task Task, actor User) ProjectActivity {
	return ProjectActivity{
		Project:      project,
		ActivityType: TaskCreated,
		Actor:        actor,
		EntityType:   TaskEntityType,
		EntityId:     task.Id,
		ActivityData: make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func ProjectUpdatedActivity(project Project, actor User) ProjectActivity {
	return ProjectActivity{
		Project:      project,
		ActivityType: ProjectUpdated,
		Actor:        actor,
		ActivityData: make(map[string]interface{}),
		EntityType:   ProjectEntityType,
		EntityId:     project.Id,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func ProjectMemberCreatedActivity(project Project, projectMember ProjectMember, actor User) ProjectActivity {
	return ProjectActivity{
		Project:      project,
		ActivityType: ProjectMemberCreated,
		Actor:        actor,
		ActivityData: make(map[string]interface{}),
		EntityType:   ProjectMemberEntityType,
		EntityId:     projectMember.Id,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TaskUpdatedActivity(project Project, task Task, actor User) ProjectActivity {
	return ProjectActivity{
		Project:      project,
		ActivityType: TaskUpdated,
		Actor:        actor,
		EntityType:   TaskEntityType,
		EntityId:     task.Id,
		ActivityData: make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TaskDeletedActivity(project Project, task Task, actor User) ProjectActivity {
	return ProjectActivity{
		Project:      project,
		ActivityType: TaskDeleted,
		Actor:        actor,
		EntityType:   TaskEntityType,
		EntityId:     task.Id,
		ActivityData: make(map[string]interface{}),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}
