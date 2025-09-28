package events

import "slices"

type Topic string

const (
	// Project
	ProjectCreated       Topic = "project.created"
	ProjectUpdated       Topic = "project.updated"
	ProjectMemberCreated Topic = "project.member.created"
	ProjectMemberRemoved Topic = "project.member.removed"

	ChatMemberCreated  Topic = "chat.member.created"
	ChatMemberViewed   Topic = "chat.member.viewed"
	ChatMessageCreated Topic = "chat.message.created"

	TaskCreated Topic = "task.created"
	TaskUpdated Topic = "task.updated"
)

func (t Topic) String() string {
	return string(t)
}

func (t Topic) Valid() bool {
	var allowedTopics = []Topic{
		ProjectCreated,
		ProjectUpdated,
		ProjectMemberCreated,
		ProjectMemberRemoved,
		ChatMemberCreated,
		ChatMessageCreated,
		ChatMemberViewed,
		TaskCreated,
		TaskUpdated,
	}

	return slices.Contains(allowedTopics, t)
}
