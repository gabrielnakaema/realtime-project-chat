package domain

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidDependencyTaskId = errors.New("depends_on_task_ids contains an invalid task id")
	ErrSelfDependency          = errors.New("task cannot depend on itself")
	ErrUnknownDependencyTasks  = errors.New("depends_on_task_ids contains unknown tasks")
	ErrDependencyCycle         = errors.New("depends_on_task_ids would create a dependency cycle")
)

func DedupeTaskIDs(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	unique := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	return unique
}

func ValidateDependencyIDs(taskId uuid.UUID, dependsOnTaskIds []uuid.UUID) error {
	for _, dependsOnTaskID := range dependsOnTaskIds {
		if dependsOnTaskID == uuid.Nil {
			return ErrInvalidDependencyTaskId
		}
		if taskId != uuid.Nil && dependsOnTaskID == taskId {
			return ErrSelfDependency
		}
	}

	return nil
}

func ResolveDependencyRefs(ids []uuid.UUID, refs []TaskDependencyRef) ([]TaskDependencyRef, error) {
	refByID := make(map[uuid.UUID]TaskDependencyRef, len(refs))
	for _, ref := range refs {
		refByID[ref.Id] = ref
	}

	orderedRefs := make([]TaskDependencyRef, 0, len(ids))
	for _, id := range ids {
		ref, ok := refByID[id]
		if !ok {
			return nil, ErrUnknownDependencyTasks
		}
		orderedRefs = append(orderedRefs, ref)
	}

	return orderedRefs, nil
}

func HasDependencyCycle(taskId uuid.UUID, dependsOnTaskIds []uuid.UUID, edges []TaskDependencyEdge) bool {
	adjacency := make(map[uuid.UUID][]uuid.UUID)
	for _, edge := range edges {
		if edge.TaskId == taskId {
			continue
		}
		adjacency[edge.TaskId] = append(adjacency[edge.TaskId], edge.DependsOnTaskId)
	}
	adjacency[taskId] = dependsOnTaskIds

	visited := map[uuid.UUID]bool{}
	for _, dependsOnTaskID := range dependsOnTaskIds {
		if dependencyPathReachesTask(dependsOnTaskID, taskId, adjacency, visited) {
			return true
		}
	}

	return false
}

func dependencyPathReachesTask(current, target uuid.UUID, adjacency map[uuid.UUID][]uuid.UUID, visited map[uuid.UUID]bool) bool {
	if current == target {
		return true
	}
	if visited[current] {
		return false
	}

	visited[current] = true
	for _, next := range adjacency[current] {
		if dependencyPathReachesTask(next, target, adjacency, visited) {
			return true
		}
	}

	return false
}
