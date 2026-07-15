package subscriber

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type realtimeNotifierStub struct {
	message       *domain.ChatMessage
	read          *domain.ChatMessageRead
	memberCreated *domain.ProjectMember
	memberRemoved *domain.ProjectMember
	roomID        uuid.UUID
}

func (s *realtimeNotifierStub) SendMessages(_ context.Context, message *domain.ChatMessage) error {
	s.message = message
	return nil
}

func (s *realtimeNotifierStub) SendReadUpdate(_ context.Context, roomID uuid.UUID, read *domain.ChatMessageRead) error {
	s.roomID = roomID
	s.read = read
	return nil
}

func (s *realtimeNotifierStub) SendChatProjectMemberCreated(_ context.Context, member *domain.ProjectMember, roomID uuid.UUID) error {
	s.memberCreated = member
	s.roomID = roomID
	return nil
}

func (s *realtimeNotifierStub) SendChatProjectMemberRemoved(_ context.Context, member *domain.ProjectMember, roomID uuid.UUID) error {
	s.memberRemoved = member
	s.roomID = roomID
	return nil
}

func TestRealtimeSubscriberDeliversEventsToWebsocketNotifier(t *testing.T) {
	chatID := uuid.New()
	projectID := uuid.New()
	userID := uuid.New()
	notifier := &realtimeNotifierStub{}
	sub := &RealtimeSubscriber{logger: slog.Default(), notifier: notifier}

	messagePayload := &events.ChatMessageCreatedPayload{ChatMessage: domain.ChatMessage{ChatId: chatID}}
	messageValue, err := json.Marshal(messagePayload)
	require.NoError(t, err)
	require.NoError(t, sub.handleEvents(context.Background(), Message{Topic: events.ChatMessageCreated, Value: messageValue}))
	require.Equal(t, chatID, notifier.message.ChatId)

	memberPayload := &events.ChatMemberCreatedPayload{
		ChatMember:    domain.ChatMember{ChatId: chatID, UserId: userID},
		ProjectMember: domain.ProjectMember{ProjectId: projectID, UserId: userID},
	}
	memberValue, err := json.Marshal(memberPayload)
	require.NoError(t, err)
	require.NoError(t, sub.handleEvents(context.Background(), Message{Topic: events.ChatMemberCreated, Value: memberValue}))
	require.Equal(t, chatID, notifier.roomID)
	require.Equal(t, projectID, notifier.memberCreated.ProjectId)

	readPayload := &events.ChatMessageReadPayload{ChatID: chatID, Read: domain.ChatMessageRead{UserId: userID}}
	readValue, err := json.Marshal(readPayload)
	require.NoError(t, err)
	require.NoError(t, sub.handleEvents(context.Background(), Message{Topic: events.ChatMessageRead, Value: readValue}))
	require.Equal(t, chatID, notifier.roomID)
	require.Equal(t, userID, notifier.read.UserId)

	removedPayload := &events.ChatMemberRemovedPayload{
		ChatMember:    domain.ChatMember{ChatId: chatID, UserId: userID},
		ProjectMember: domain.ProjectMember{ProjectId: projectID, UserId: userID},
	}
	removedValue, err := json.Marshal(removedPayload)
	require.NoError(t, err)
	require.NoError(t, sub.handleEvents(context.Background(), Message{Topic: events.ChatMemberRemoved, Value: removedValue}))
	require.Equal(t, chatID, notifier.roomID)
	require.Equal(t, projectID, notifier.memberRemoved.ProjectId)
}
