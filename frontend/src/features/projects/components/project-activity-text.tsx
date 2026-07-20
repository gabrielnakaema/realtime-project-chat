import type {
  ProjectActivity,
  ProjectMemberProjectActivity,
  ProjectProjectActivity,
  TaskProjectActivity,
} from '@/features/projects/types/project-activity';
import { ActionOriginBadge } from '@/features/tasks/components/action-origin-badge';

export const ProjectActivityText = ({ activity }: { activity: ProjectActivity }) => {
  if (activity.entity_type === 'task') {
    switch (activity.activity_type) {
      case 'task.created':
        return <TaskCreatedActivityText activity={activity} />;
      case 'task.updated':
        return <TaskUpdatedActivityText activity={activity} />;
      case 'task.deleted':
        return <TaskDeletedActivityText activity={activity} />;
      default:
        return <GenericActivityText activity={activity} />;
    }
  }

  if (activity.entity_type === 'project') {
    switch (activity.activity_type) {
      case 'project.created':
        return <ProjectCreatedActivityText activity={activity} />;
      case 'project.updated':
        return <ProjectUpdatedActivityText activity={activity} />;
      default:
        return <GenericActivityText activity={activity} />;
    }
  }

  if (activity.entity_type === 'project.member') {
    switch (activity.activity_type) {
      case 'project.member.created':
        return <ProjectMemberCreatedActivityText activity={activity} />;
      case 'project.member.deleted':
        return <ProjectMemberDeletedActivityText activity={activity} />;
      default:
        return <GenericActivityText activity={activity} />;
    }
  }

  return <GenericActivityText activity={activity} />;
};

const GeneralText = ({ children }: { children: React.ReactNode }) => {
  return <p className="text-foreground text-sm">{children}</p>;
};

const Actor = ({ children }: { children: React.ReactNode }) => {
  return <span className="font-semibold">{children}</span>;
};

const Subject = ({ children }: { children: React.ReactNode }) => {
  return <span className="font-semibold">{children}</span>;
};

const GenericActivityText = ({ activity }: { activity: ProjectActivity }) => {
  return (
    <GeneralText>
      <ActorWithOrigin name={activity.actor?.name} actionOrigin={activity.action_origin} /> performed an action
    </GeneralText>
  );
};

const TaskCreatedActivityText = ({ activity }: { activity: TaskProjectActivity }) => {
  return (
    <GeneralText>
      <ActorWithOrigin name={activity.actor?.name} actionOrigin={activity.action_origin} /> created task{' '}
      <Subject>{activity.task?.title}</Subject>
    </GeneralText>
  );
};

const TaskUpdatedActivityText = ({ activity }: { activity: TaskProjectActivity }) => {
  return (
    <GeneralText>
      <ActorWithOrigin name={activity.actor?.name} actionOrigin={activity.action_origin} /> updated task{' '}
      <Subject>{activity.task?.title}</Subject>
    </GeneralText>
  );
};

const TaskDeletedActivityText = ({ activity }: { activity: TaskProjectActivity }) => {
  return (
    <GeneralText>
      <ActorWithOrigin name={activity.actor?.name} actionOrigin={activity.action_origin} /> deleted task{' '}
      <Subject>{activity.task?.title}</Subject>
    </GeneralText>
  );
};

const ProjectCreatedActivityText = ({ activity }: { activity: ProjectProjectActivity }) => {
  return (
    <GeneralText>
      <ActorWithOrigin name={activity.actor?.name} actionOrigin={activity.action_origin} /> created this project
    </GeneralText>
  );
};

const ProjectUpdatedActivityText = ({ activity }: { activity: ProjectProjectActivity }) => {
  return (
    <GeneralText>
      <ActorWithOrigin name={activity.actor?.name} actionOrigin={activity.action_origin} /> updated this project
    </GeneralText>
  );
};

const ProjectMemberCreatedActivityText = ({ activity }: { activity: ProjectMemberProjectActivity }) => {
  return (
    <GeneralText>
      <ActorWithOrigin name={activity.actor?.name} actionOrigin={activity.action_origin} /> added a new member to the
      project
    </GeneralText>
  );
};

const ProjectMemberDeletedActivityText = ({ activity }: { activity: ProjectMemberProjectActivity }) => {
  return (
    <GeneralText>
      <ActorWithOrigin name={activity.actor?.name} actionOrigin={activity.action_origin} /> removed a member from the
      project
    </GeneralText>
  );
};

const ActorWithOrigin = ({
  name,
  actionOrigin,
}: {
  name?: React.ReactNode;
  actionOrigin?: ProjectActivity['action_origin'];
}) => {
  return (
    <span className="inline-flex flex-wrap items-center gap-1.5 align-middle">
      <Actor>{name ?? 'Someone'}</Actor>
      <ActionOriginBadge origin={actionOrigin} />
    </span>
  );
};
