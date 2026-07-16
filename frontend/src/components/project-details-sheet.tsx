import { Calendar, Crown, Loader2, Users } from 'lucide-react';
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from './ui/sheet';
import { ProjectActivity } from './project-activity';
import type { Project } from '@/types/project';
import { richTextListClassName, sanitizeHTML } from '@/utils/html';
import { cn } from '@/lib/utils';
import { formatLongDate } from '@/utils/date';
import { ProjectMemberRole } from '@/types/project';
import { useProjectActivities } from '@/hooks/use-project-activities';

interface ProjectDetailsSheetProps {
  project: Project;
  children: React.ReactNode;
}

export const ProjectDetailsSheetTrigger = SheetTrigger;

export const ProjectDetailsSheet = ({ project, children }: ProjectDetailsSheetProps) => {
  const creator = project.members.find((m) => m.role === ProjectMemberRole.Creator);
  const formattedDate = formatLongDate(project.created_at);

  const { data: activities, queryResult, sentinelRef } = useProjectActivities(project.id);
  const { isFetchingNextPage } = queryResult;

  return (
    <Sheet>
      {children}
      <SheetContent>
        <SheetHeader>
          <SheetTitle>{project.name}</SheetTitle>
        </SheetHeader>

        <div className="flex flex-col gap-6 overflow-y-auto p-6">
          <div className="flex flex-col gap-3">
            <div className="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400">
              <Calendar className="h-4 w-4" />
              <span>Created on {formattedDate}</span>
            </div>

            {creator && (
              <div className="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400">
                <Crown className="h-4 w-4" />
                <span>Created by {creator.user.name}</span>
              </div>
            )}

            <div className="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400">
              <Users className="h-4 w-4" />
              <span>
                {project.members.length} {project.members.length === 1 ? 'member' : 'members'}
              </span>
            </div>
          </div>

          {project.description && (
            <div className="flex flex-col gap-2">
              <h3 className="text-sm font-medium text-slate-900 dark:text-slate-100">Description</h3>
              <div
                className={cn('text-sm text-slate-700 dark:text-slate-300', richTextListClassName)}
                dangerouslySetInnerHTML={{ __html: sanitizeHTML(project.description) }}
              />
            </div>
          )}

          {activities.length > 0 && (
            <div className="flex flex-col gap-2">
              <h3 className="text-sm font-medium text-slate-900 dark:text-slate-100">Activity</h3>
              <div className="flex flex-col gap-4 overflow-y-auto pr-1">
                {activities.map((activity) => (
                  <ProjectActivity key={activity.id} activity={activity} showProject={false} />
                ))}
                <div ref={sentinelRef} className="h-1" aria-hidden="true" />
                {isFetchingNextPage && (
                  <div className="flex justify-center py-2">
                    <Loader2 className="h-4 w-4 animate-spin text-slate-400" />
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
};
