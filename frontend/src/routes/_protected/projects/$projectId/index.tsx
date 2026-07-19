import { createFileRoute } from '@tanstack/react-router';
import { RotateCcw } from 'lucide-react';

import { z } from 'zod';

import { KanbanBoard } from '@/features/tasks/components/kanban-board';

import { TaskDetails } from '@/features/tasks/components/task-details';
import { EditTask } from '@/features/tasks/components/task-form/edit-task';

import { useProjectDetails } from '@/features/projects/hooks/use-project-details';
import { useTaskDetailsRouting } from '@/features/tasks/hooks/use-task-details-routing';

import { ProjectDetailsHeader } from '@/features/projects/components/project-details-page/project-details-header';
import { Button } from '@/shared/components/button';

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

  return <ProjectDetailsPage projectId={projectId} />;
}

export const ProjectDetailsPage = ({ projectId }: { projectId: string }) => {
  const { data: project, isError, isLoading, refetch } = useProjectDetails(projectId);
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
      {isLoading && <ProjectDetailsSkeleton />}

      {!isLoading && isError && <ProjectDetailsError onRetry={() => void refetch()} />}

      {!isLoading && !isError && !project && (
        <div className="flex flex-1 items-center justify-center p-6">
          <p className="text-muted-foreground text-sm">Project not found.</p>
        </div>
      )}

      {!isLoading && !isError && project && (
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
};

const ProjectDetailsError = ({ onRetry }: { onRetry: () => void }) => (
  <div className="flex flex-1 items-center justify-center p-6">
    <div className="border-border bg-card flex min-h-48 w-full max-w-lg flex-col items-center justify-center rounded-xl border p-8 text-center">
      <p className="text-foreground text-sm font-medium">Project could not be loaded.</p>
      <p className="text-muted-foreground mt-1 text-xs">Check your connection and try again.</p>
      <Button type="button" variant="outline" size="sm" className="mt-4" onClick={onRetry}>
        <RotateCcw className="size-3.5" />
        Try again
      </Button>
    </div>
  </div>
);

const ProjectDetailsSkeleton = () => (
  <div className="flex min-h-screen animate-pulse flex-col" aria-label="Loading project">
    <div className="border-border bg-card flex h-16 shrink-0 items-center justify-between border-b px-4 sm:px-6">
      <div className="flex items-center gap-3">
        <div className="bg-muted size-9 rounded-md" />
        <div className="bg-muted h-9 w-40 rounded-md" />
      </div>
      <div className="flex items-center gap-2">
        <div className="bg-muted h-9 w-28 rounded-md" />
        <div className="bg-muted size-9 rounded-md" />
        <div className="bg-muted size-9 rounded-md" />
      </div>
    </div>

    <div className="bg-background flex flex-1 gap-6 overflow-hidden p-6">
      {Array.from({ length: 4 }).map((_, index) => (
        <div key={index} className="bg-muted/40 h-[calc(100vh-100px)] w-80 shrink-0 rounded-lg p-3">
          <div className="mb-4 flex items-center justify-between">
            <div className="bg-muted h-4 w-28 rounded" />
            <div className="bg-muted size-5 rounded" />
          </div>
          <div className="space-y-3">
            {Array.from({ length: index % 2 === 0 ? 3 : 2 }).map((__, cardIndex) => (
              <div key={cardIndex} className="border-border bg-card h-28 rounded-md border" />
            ))}
          </div>
        </div>
      ))}
    </div>
  </div>
);
