import { Link } from '@tanstack/react-router';
import { ChevronDown } from 'lucide-react';
import { LoadingSpinner } from '../../../../components/loading';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useListProjects } from '@/hooks/use-list-projects';
import { Button } from '@/shared/components/button';

interface ProjectsDropdownProps {
  current: {
    value: string;
    label: string;
  };
}

export const ProjectsDropdown = ({ current }: ProjectsDropdownProps) => {
  const { data: projects, isLoading } = useListProjects();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button type="button" variant="ghost">
          <span>{current.label}</span>

          <ChevronDown className="text-muted-foreground h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start">
        {isLoading && <LoadingSpinner size="2rem" />}
        {projects?.map((project) => (
          <DropdownMenuItem key={project.id} asChild>
            <Link
              to="/projects/$projectId"
              params={{
                projectId: project.id,
              }}
            >
              {project.name}
            </Link>
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
};
