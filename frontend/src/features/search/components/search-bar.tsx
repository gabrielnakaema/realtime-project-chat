import { useState } from 'react';
import { Search } from 'lucide-react';
import type { FormEvent } from 'react';
import { Button } from '@/shared/components/button';
import { Input } from '@/shared/components/input';
import { cn } from '@/lib/utils';
import { hasSearchQuery, normalizeSearchQuery } from '@/shared/utils/search';

interface SearchBarProps {
  action: string;
  searchName: string;
  formClassName?: string;
  inputClassName?: string;
  initialValue?: string;
}

export const SearchBar = ({ action, searchName, formClassName, initialValue }: SearchBarProps) => {
  const [value, setValue] = useState(initialValue ?? '');

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    const normalizedValue = normalizeSearchQuery(value);
    if (!normalizedValue) {
      event.preventDefault();
      return;
    }

    if (normalizedValue !== value) {
      const input = event.currentTarget.elements.namedItem(searchName);
      if (input instanceof HTMLInputElement) {
        input.value = normalizedValue;
      }
      setValue(normalizedValue);
    }
  };

  return (
    <form action={action} method="GET" className={cn('flex items-center gap-2', formClassName)} onSubmit={handleSubmit}>
      <label htmlFor={searchName} className="sr-only">
        Search projects and tasks
      </label>
      <Input
        id={searchName}
        label=""
        type="search"
        name={searchName}
        placeholder="Search projects and tasks"
        required
        value={value}
        onChange={(event) => setValue(event.target.value)}
      />
      <Button type="submit" className="p-3" disabled={!hasSearchQuery(value)}>
        <Search className="h-4 w-4" />
      </Button>
    </form>
  );
};
