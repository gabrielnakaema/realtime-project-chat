import { Search } from 'lucide-react';
import { Button } from './button';
import { Input } from './input';
import { cn } from '@/lib/utils';

interface SearchBarProps {
  action: string;
  searchName: string;
  formClassName?: string;
  inputClassName?: string;
  initialValue?: string;
}

export const SearchBar = ({ action, searchName, formClassName, initialValue }: SearchBarProps) => {
  return (
    <form action={action} method="GET" className={cn('flex items-center gap-2', formClassName)}>
      <label htmlFor={searchName} className="sr-only">
        Search projects and tasks
      </label>
      <Input
        id={searchName}
        label=""
        type="search"
        name={searchName}
        placeholder="Search"
        required
        defaultValue={initialValue}
      />
      <Button type="submit" className="p-3">
        <Search className="h-4 w-4" />
      </Button>
    </form>
  );
};
