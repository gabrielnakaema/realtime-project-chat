import { CheckIcon, PencilIcon, PlusIcon, TrashIcon, UserMinusIcon, UserPlusIcon } from 'lucide-react';
import { Link } from '@tanstack/react-router';
import { ProjectActivityText } from './project-activity-text';
import type { ProjectActivity as ProjectActivityType } from '@/types/project-activity';
import { formatRelativeActivityDateString } from '@/utils/format-relative-activity';

export const ProjectActivity = ({
  activity,
  showProject = true,
}: {
  activity: ProjectActivityType;
  showProject?: boolean;
}) => {
  return (
    <div className="flex items-center gap-4">
      {activityIcons[activity.activity_type]}
      <div className="flex flex-col gap-1">
        <ProjectActivityText activity={activity} />
        <p className="text-muted-foreground text-xs">
          {showProject && activity.project && (
            <>
              <Link to="/projects/$projectId" params={{ projectId: activity.project.id }} className="hover:underline">
                {activity.project.name}
              </Link>
              {' · '}
            </>
          )}
          {formatRelativeActivityDateString(activity.created_at)}
        </p>
      </div>
    </div>
  );
};

const activityIcons: Partial<Record<ProjectActivityType['activity_type'], React.ReactNode>> = {
  'project.member.created': <UserPlusIcon className="h-5 w-5" />,
  'project.member.deleted': <UserMinusIcon className="h-5 w-5" />,
  'task.created': <CheckIcon className="h-5 w-5" />,
  'task.updated': <PencilIcon className="h-5 w-5" />,
  'task.deleted': <TrashIcon className="h-5 w-5" />,
  'project.created': <PlusIcon className="h-5 w-5" />,
  'project.updated': <PencilIcon className="h-5 w-5" />,
};
