import { CheckIcon, PencilIcon, PlusIcon, TrashIcon, UserMinusIcon, UserPlusIcon } from 'lucide-react';
import { ProjectActivityText } from './project-activity-text';
import type { ProjectActivity as ProjectActivityType } from '@/types/project-activity';
import { formatRelativeActivityDateString } from '@/utils/format-relative-activity';

export const ProjectActivity = ({ activity }: { activity: ProjectActivityType }) => {
  return (
    <div className="flex items-center gap-4">
      {activityIcons[activity.activity_type]}
      <div className="flex flex-col gap-1">
        <ProjectActivityText activity={activity} />
        <p className="text-xs text-slate-500 dark:text-slate-400">
          {formatRelativeActivityDateString(activity.created_at)}
        </p>
      </div>
    </div>
  );
};

const activityIcons: Partial<Record<ProjectActivityType['activity_type'], React.ReactNode>> = {
  'project.member.created': <UserPlusIcon className="w-5 h-5" />,
  'project.member.deleted': <UserMinusIcon className="w-5 h-5" />,
  'task.created': <CheckIcon className="w-5 h-5" />,
  'task.updated': <PencilIcon className="w-5 h-5" />,
  'task.deleted': <TrashIcon className="w-5 h-5" />,
  'project.created': <PlusIcon className="w-5 h-5" />,
  'project.updated': <PencilIcon className="w-5 h-5" />,
};
