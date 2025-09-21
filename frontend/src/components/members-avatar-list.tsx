import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip';

interface MembersAvatarListProps {
  names: string[];
  max?: number;
}

export const MembersAvatarList = ({ names, max = 4 }: MembersAvatarListProps) => {
  const namesToShow = names.slice(0, max);
  const remainingNames = names.slice(max);
  const remaining = names.length - max;

  return (
    <div className="flex -space-x-2">
      {namesToShow.map((name, index) => (
        <Tooltip key={`${name}-${index}`}>
          <TooltipTrigger>
            <div className="w-8 h-8 bg-blue-600 rounded-full border-2 border-white dark:border-slate-800 flex items-center justify-center text-white text-xs font-medium">
              {name.charAt(0).toUpperCase()}
            </div>
          </TooltipTrigger>
          <TooltipContent>{name}</TooltipContent>
        </Tooltip>
      ))}
      {remaining > 0 && (
        <Tooltip>
          <TooltipTrigger>
            <div className="w-8 h-8 rounded-full bg-slate-200 dark:bg-slate-700 border-2 border-white dark:border-slate-800 flex items-center justify-center">
              <span className="text-xs font-medium text-slate-600 dark:text-slate-400">+{remaining}</span>
            </div>
          </TooltipTrigger>
          <TooltipContent>
            {remainingNames.map((name) => (
              <p key={name}>{name}</p>
            ))}
          </TooltipContent>
        </Tooltip>
      )}
    </div>
  );
};
