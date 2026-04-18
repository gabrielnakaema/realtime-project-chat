import { Link } from '@tanstack/react-router';
import { ArrowLeft } from 'lucide-react';
import type { Project } from '@/types/project';
import { AddProjectMember } from '@/components/add-project-member';
import { MembersAvatarList } from '@/components/members-avatar-list';
import { ProjectMembersModal } from '@/components/project-members-modal';

interface ProjectChatHeaderProps {
  projectId: string;
  project?: Project;
  onlineUserIds: string[];
}

export const ProjectChatHeader = ({ projectId, project, onlineUserIds }: ProjectChatHeaderProps) => {
  return (
    <header className="border-b border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
      <div className="px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link
              to="/projects/$projectId"
              className="inline-flex items-center rounded-md px-3 py-2 font-medium text-slate-700 transition-colors hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-700"
              params={{ projectId }}
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Go back
            </Link>
            <div>
              <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">Team chat - {project?.name}</h1>
              <p className="text-slate-600 dark:text-slate-400">Chat with your team members</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <AddProjectMember projectId={projectId} />
            <ProjectMembersModal project={project} />
            <MembersAvatarList
              onlineUserIds={onlineUserIds}
              members={
                project?.members.map((member) => ({
                  user_id: member.user_id,
                  name: member.user.name || '',
                })) || []
              }
              max={4}
            />
          </div>
        </div>
      </div>
    </header>
  );
};
