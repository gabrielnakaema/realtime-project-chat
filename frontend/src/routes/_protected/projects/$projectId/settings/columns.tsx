import { createFileRoute } from '@tanstack/react-router';
import { ColumnsProjectSettings } from '@/features/projects/components/project-settings/columns-project-settings';

export const Route = createFileRoute('/_protected/projects/$projectId/settings/columns')({
  component: RouteComponent,
});

function RouteComponent() {
  const params = Route.useParams();

  return <ColumnsProjectSettings projectId={params.projectId} />;
}
