interface TagPreviewListProps {
  tags: string[];
}

export const TagPreviewList = ({ tags }: TagPreviewListProps) => {
  if (tags.length === 0) {
    return null;
  }

  return (
    <div className="flex w-full flex-wrap gap-2">
      {tags.map((tag) => (
        <div
          key={tag}
          className="border-border text-muted-foreground w-fit rounded-sm border px-2 py-0.5 text-xs font-medium"
        >
          {tag}
        </div>
      ))}
    </div>
  );
};
