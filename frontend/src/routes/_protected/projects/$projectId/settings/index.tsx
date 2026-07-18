import { createFileRoute } from '@tanstack/react-router';
import { GeneralProjectSettingsForm } from '@/features/projects/components/project-settings/general-project-settings-form';

export const Route = createFileRoute('/_protected/projects/$projectId/settings/')({
  component: RouteComponent,
});

function RouteComponent() {
  const params = Route.useParams();
  return <GeneralProjectSettingsForm projectId={params.projectId} />;
}
