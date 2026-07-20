import { Package } from 'lucide-react';
import { ProjectCard, ProjectCardSkeleton } from '@/features/projects/components/project-card';
import { SearchEmptyState } from '@/features/search/components/empty-state';
import { useSearchProjects } from '@/features/search/hooks/use-search-projects';

interface ProjectSearchResultsProps {
  query?: string;
}

export const ProjectSearchResults = ({ query }: ProjectSearchResultsProps) => {
  const { data: projects, isLoading } = useSearchProjects(query);
  const areProjectsEmpty = !isLoading && !projects?.length;

  return (
    <section className="flex flex-col gap-4">
      <h2 className="text-foreground text-xl font-bold">Projects</h2>
      <div className="flex w-full max-w-full gap-4 overflow-x-auto overflow-y-auto">
        {isLoading && <ProjectSearchSkeleton />}
        {areProjectsEmpty && (
          <SearchEmptyState
            icon={<Package className="h-4 w-4" />}
            title="No projects found"
            description="There are no projects for your search. Try adjusting your search terms."
          />
        )}
        {projects?.map((project) => (
          <ProjectCard key={project.id} project={project} />
        ))}
      </div>
    </section>
  );
};

const ProjectSearchSkeleton = () => {
  return Array.from({ length: 5 }).map((_, index) => <ProjectCardSkeleton key={index} />);
};
