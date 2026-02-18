import { Calendar, Crown, Users } from 'lucide-react';
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from './ui/sheet';
import type { Project } from '@/types/project';
import { sanitizeHTML } from '@/utils/html';
import { ProjectMemberRole } from '@/types/project';

interface ProjectDetailsSheetProps {
  project: Project;
  children: React.ReactNode;
}

export const ProjectDetailsSheetTrigger = SheetTrigger;

export const ProjectDetailsSheet = ({ project, children }: ProjectDetailsSheetProps) => {
  const creator = project.members.find((m) => m.role === ProjectMemberRole.Creator);
  const formattedDate = new Date(project.created_at).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });

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
                className="text-sm text-slate-700 dark:text-slate-300"
                dangerouslySetInnerHTML={{ __html: sanitizeHTML(project.description) }}
              />
            </div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
};
