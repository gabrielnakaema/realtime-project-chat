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
          className="w-fit rounded-sm border border-slate-200 px-2 py-0.5 text-xs font-medium text-slate-500 dark:border-slate-700 dark:text-slate-400"
        >
          {tag}
        </div>
      ))}
    </div>
  );
};
