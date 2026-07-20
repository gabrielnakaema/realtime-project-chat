package tasks_test

import (
	"context"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockProjectRepository struct {
	mock.Mock
}

func (m *mockProjectRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *mockProjectRepository) GetColumnById(ctx context.Context, id uuid.UUID) (*domain.ProjectColumn, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProjectColumn), args.Error(1)
}

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
