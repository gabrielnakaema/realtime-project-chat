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
    <div className="relative w-full space-y-[6px]">
      <label htmlFor={id} className={cn('block text-[11px] font-semibold tracking-wider')}>
        {label}
      </label>
      <SelectPrimitive value={value} onValueChange={onChange}>
        <SelectTrigger className="border-border bg-card text-foreground focus:ring-ring [&[data-placeholder]]:text-muted-foreground h-[39px] w-full rounded-md border px-3 text-base focus:border-transparent focus:ring-2 focus:outline-none">
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
      {error && <p className="text-destructive text-xs">{error}</p>}
    </div>
  );
};
