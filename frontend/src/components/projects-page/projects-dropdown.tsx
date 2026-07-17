import { Link } from '@tanstack/react-router';
import { ChevronDown } from 'lucide-react';
import { LoadingSpinner } from '../loading';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useListProjects } from '@/hooks/use-list-projects';

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
        <button
          type="button"
          className="dark:text-foreground flex items-center gap-2 rounded-sm p-2 pl-3 font-sans text-[13px] font-medium tracking-[-0.01em]"
        >
          <span>{current.label}</span>

          <ChevronDown className="h-4 w-4 text-slate-500" />
        </button>
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
