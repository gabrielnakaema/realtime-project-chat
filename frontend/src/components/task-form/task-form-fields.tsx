import { Controller } from 'react-hook-form';
import { Input } from '../input';
import { Select } from '../select';
import { TextEditor } from '../text-editor';
import { TagPreviewList } from './tag-preview-list';
import { parseUniqueTags, priorityOptions } from './task-form-utils';
import type { Control, FieldErrors, UseFormRegister, UseFormSetValue } from 'react-hook-form';
import type { ITaskForm } from '@/schemas/task-schema';

interface TaskMemberOption {
  label: string;
  value: string;
}

interface TaskFormFieldsProps {
  control: Control<ITaskForm>;
  register: UseFormRegister<ITaskForm>;
  setValue: UseFormSetValue<ITaskForm>;
  errors: FieldErrors<ITaskForm>;
  memberOptions: TaskMemberOption[];
  descriptionInitialValue: string;
  descriptionEditorKey?: string;
}

export const TaskFormFields = ({
  control,
  register,
  setValue,
  errors,
  memberOptions,
  descriptionInitialValue,
  descriptionEditorKey,
}: TaskFormFieldsProps) => {
  return (
    <>
      <Input
        label="Title"
        id="title"
        placeholder="Enter task title"
        {...register('title')}
        error={errors.title?.message}
      />
      <Controller
        control={control}
        name="responsible_id"
        render={({ field }) => (
          <Select
            options={memberOptions}
            value={field.value ?? ''}
            onChange={field.onChange}
            label="Responsible"
            error={errors.responsible_id?.message}
            id="responsible_id"
            placeholder="Select responsible"
          />
        )}
      />
      <div className="flex w-full gap-4">
        <Controller
          control={control}
          name="priority"
          render={({ field }) => (
            <Select
              options={priorityOptions}
              value={field.value}
              onChange={field.onChange}
              label="Priority"
              error={errors.priority?.message}
              id="priority"
              placeholder="Select priority"
            />
          )}
        />
        <Input
          label="Due date"
          id="due_date"
          type="date"
          placeholder="Enter due date"
          {...register('due_date')}
          error={errors.due_date?.message}
        />
      </div>
      <Controller
        control={control}
        name="tags"
        render={({ field }) => {
          const tags = parseUniqueTags(field.value);

          return (
            <>
              <Input
                label="Tags"
                id="tags"
                placeholder="Enter comma separated tags"
                value={field.value ?? ''}
                onChange={field.onChange}
                ref={field.ref}
                error={errors.tags?.message}
              />
              <TagPreviewList tags={tags} />
            </>
          );
        }}
      />
      <TextEditor
        key={descriptionEditorKey}
        initialValue={descriptionInitialValue}
        onChange={(html) => setValue('description', html)}
        label="Description"
        id="description"
        placeholder="Enter task description"
        error={errors.description?.message}
      />
    </>
  );
};
