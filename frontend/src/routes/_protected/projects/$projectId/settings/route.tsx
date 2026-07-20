import { createFileRoute } from '@tanstack/react-router';
import { ProjectSettingsLayout } from '@/features/projects/components/project-settings/project-settings-layout';

export const Route = createFileRoute('/_protected/projects/$projectId/settings')({
  component: RouteComponent,
});

function RouteComponent() {
  const { projectId } = Route.useParams();

  return <ProjectSettingsLayout projectId={projectId} />;
}
