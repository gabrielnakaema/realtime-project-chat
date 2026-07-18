import { createFileRoute } from '@tanstack/react-router';
import { MembersProjectSettings } from '@/features/projects/components/project-settings/members-project-settings';

export const Route = createFileRoute('/_protected/projects/$projectId/settings/members')({
  component: RouteComponent,
});

function RouteComponent() {
  const params = Route.useParams();

  return <MembersProjectSettings projectId={params.projectId} />;
}
