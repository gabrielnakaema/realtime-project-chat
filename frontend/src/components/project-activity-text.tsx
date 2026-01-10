import type {
  ProjectActivity,
  ProjectMemberProjectActivity,
  ProjectProjectActivity,
  TaskProjectActivity,
} from '@/types/project-activity';

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
      <Actor>{activity.actor?.name}</Actor> performed an action in <Subject>{activity.project?.name}</Subject>
    </GeneralText>
  );
};

const TaskCreatedActivityText = ({ activity }: { activity: TaskProjectActivity }) => {
  return (
    <GeneralText>
      <Actor>{activity.actor?.name}</Actor> created a new task <Subject>{activity.task?.title}</Subject> in{' '}
      <Subject>{activity.project?.name}</Subject>
    </GeneralText>
  );
};

const TaskUpdatedActivityText = ({ activity }: { activity: TaskProjectActivity }) => {
  return (
    <GeneralText>
      <Actor>{activity.actor?.name}</Actor> updated the task <Subject>{activity.task?.title}</Subject> in{' '}
      <Subject>{activity.project?.name}</Subject>
    </GeneralText>
  );
};

const TaskDeletedActivityText = ({ activity }: { activity: TaskProjectActivity }) => {
  return (
    <GeneralText>
      <Actor>{activity.actor?.name}</Actor> deleted the task <Subject>{activity.task?.title}</Subject> in{' '}
      <Subject>{activity.project?.name}</Subject>
    </GeneralText>
  );
};

const ProjectCreatedActivityText = ({ activity }: { activity: ProjectProjectActivity }) => {
  return (
    <GeneralText>
      <Actor>{activity.actor?.name}</Actor> created project <Subject>{activity.project?.name}</Subject>
    </GeneralText>
  );
};

const ProjectUpdatedActivityText = ({ activity }: { activity: ProjectProjectActivity }) => {
  return (
    <GeneralText>
      <Actor>{activity.actor?.name}</Actor> updated project <Subject>{activity.project?.name}</Subject>
    </GeneralText>
  );
};

const ProjectMemberCreatedActivityText = ({ activity }: { activity: ProjectMemberProjectActivity }) => {
  return (
    <GeneralText>
      <Actor>{activity.actor?.name}</Actor> added a new member to the project{' '}
      <Subject>{activity.project?.name}</Subject>
    </GeneralText>
  );
};

const ProjectMemberDeletedActivityText = ({ activity }: { activity: ProjectMemberProjectActivity }) => {
  return (
    <GeneralText>
      <Actor>{activity.actor?.name}</Actor> removed a member from the project{' '}
      <Subject>{activity.project?.name}</Subject>
    </GeneralText>
  );
};
