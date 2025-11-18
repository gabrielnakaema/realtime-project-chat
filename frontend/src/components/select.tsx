import { SelectContent, SelectItem, Select as SelectPrimitive, SelectTrigger, SelectValue } from './ui/select';
import { cn } from '@/lib/utils';

interface SelectProps {
  options: {
    label: string;
    value: string;
    className?: string;
  }[];
  value: string;
  onChange: (value: string) => void;
  label: string;
  error?: string;
  id?: string;
  placeholder?: string;
}

export const Select = ({ options, value, onChange, label, error, id, placeholder }: SelectProps) => {
  return (
    <div className="w-full space-y-2">
      <label htmlFor={id} className="block text-sm font-medium text-slate-700 dark:text-slate-300">
        {label}
      </label>
      <SelectPrimitive value={value} onValueChange={onChange}>
        <SelectTrigger className="w-full rounded-md border border-slate-300 bg-white px-3 py-5 text-base text-slate-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-slate-100 [&[data-placeholder]]:text-slate-500 dark:[&[data-placeholder]]:text-slate-400">
          <SelectValue id={id} placeholder={placeholder} className="text-base focus:ring-0" />
        </SelectTrigger>
        <SelectContent>
          {options.map((option) => (
            <SelectItem key={option.value} value={option.value} className={cn('text-base', option.className)}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </SelectPrimitive>
      {error && <p className="text-sm text-red-500">{error}</p>}
    </div>
  );
};
