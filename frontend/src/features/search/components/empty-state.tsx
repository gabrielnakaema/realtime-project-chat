interface EmptyStateProps {
  icon: React.ReactNode;
  title: string;
  description: string;
}

export const SearchEmptyState = ({ icon, title, description }: EmptyStateProps) => {
  return (
    <div className="border-border flex w-full flex-col items-center justify-center rounded-lg border border-dashed px-6 py-12 text-center">
      <div className="bg-secondary text-muted-foreground mb-4 flex h-12 w-12 items-center justify-center rounded-full">
        {icon}
      </div>
      <h3 className="text-foreground mb-1 text-sm font-medium">{title}</h3>
      <p className="text-muted-foreground mb-4 max-w-sm text-sm">{description}</p>
    </div>
  );
};
