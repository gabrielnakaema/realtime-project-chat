package events

import "slices"

type Topic string

const (
	ProjectCreated       Topic = "project.created"
	ProjectUpdated       Topic = "project.updated"
	ProjectDeleted       Topic = "project.deleted"
	ProjectMemberCreated Topic = "project.member.created"
	ProjectMemberRemoved Topic = "project.member.removed"

	ChatMemberCreated  Topic = "chat.member.created"
	ChatMemberRemoved  Topic = "chat.member.removed"
	ChatMemberViewed   Topic = "chat.member.viewed"
	ChatMessageCreated Topic = "chat.message.created"
	ChatMessageRead    Topic = "chat.message.read"

	TaskCreated        Topic = "task.created"
	TaskUpdated        Topic = "task.updated"
	TaskCommentCreated Topic = "task.comment.created"

	NotificationCreated Topic = "notification.created"
)

func (t Topic) String() string {
	return string(t)
}

func (t Topic) Valid() bool {
	var allowedTopics = []Topic{
		ProjectCreated,
		ProjectUpdated,
		ProjectDeleted,
		ProjectMemberCreated,
		ProjectMemberRemoved,
		ChatMemberCreated,
		ChatMemberRemoved,
		ChatMessageCreated,
		ChatMemberViewed,
		ChatMessageRead,
		TaskCreated,
		TaskUpdated,
		TaskCommentCreated,
		NotificationCreated,
	}

	return slices.Contains(allowedTopics, t)
}
