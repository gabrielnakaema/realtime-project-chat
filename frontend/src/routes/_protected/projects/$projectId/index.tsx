import { createFileRoute } from '@tanstack/react-router';

import { z } from 'zod';

import { KanbanBoard } from '@/components/kanban-board';

import { TaskDetails } from '@/components/task-details';
import { EditTask } from '@/components/task-form/edit-task';

import { useProjectDetails } from '@/hooks/use-project-details';
import { useTaskDetailsRouting } from '@/hooks/use-task-details-routing';

import { ProjectDetailsHeader } from '@/features/projects/components/project-details-page/project-details-header';

export const Route = createFileRoute('/_protected/projects/$projectId/')({
  component: RouteComponent,
  validateSearch: z.object({
    taskId: z.string().optional(),
    commentId: z.string().optional(),
    commentCreatedAt: z.string().optional(),
  }),
});

function RouteComponent() {
  const { projectId } = Route.useParams();
  const { data: project } = useProjectDetails(projectId);
  const {
    selectedTaskId,
    selectedCommentId,
    selectedCommentCreatedAt,
    isEditingTask,
    closeTask,
    startEditingTask,
    stopEditingTask,
  } = useTaskDetailsRouting();

  return (
    <div className="bg-muted flex h-fit min-h-screen flex-col">
      {project && (
        <>
          <ProjectDetailsHeader project={project} />
          <div className="bg-background flex-1 p-6">
            <KanbanBoard project={project} />
          </div>
        </>
      )}

      <TaskDetails
        taskId={selectedTaskId}
        open={!!selectedTaskId && !isEditingTask}
        targetCommentId={selectedCommentId}
        targetCommentCreatedAt={selectedCommentCreatedAt}
        onOpenChange={(open) => {
          if (!open) {
            closeTask();
          }
        }}
        onEdit={startEditingTask}
      />

      <EditTask onOpenChange={stopEditingTask} taskId={selectedTaskId} open={isEditingTask} />
    </div>
  );
}
