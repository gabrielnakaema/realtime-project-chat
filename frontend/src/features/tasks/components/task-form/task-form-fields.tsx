import { Controller, useFormContext } from 'react-hook-form';
import { TaskCodeAutocomplete } from './task-code-autocomplete';
import { TagPreviewList } from './tag-preview-list';
import { TaskDependencyPicker } from './task-dependency-picker';
import { parseUniqueTags, priorityOptions } from './task-form-utils';
import type { ITaskForm } from '@/features/tasks/schemas/task.schema';
import type { TaskMemberOption } from './task-form-utils';
import type { TaskDependencyRef } from '@/features/tasks/types/task';
import { TextEditor } from '@/shared/components/text-editor';
import { Select } from '@/shared/components/select';
import { Input } from '@/shared/components/input';

interface TaskFormFieldsProps {
  projectId: string;
  excludeTaskId?: string;
  initialDependencyTasks?: TaskDependencyRef[];
  memberOptions: TaskMemberOption[];
  descriptionInitialValue: string;
  descriptionEditorKey?: string;
}

export const TaskFormFields = ({
  projectId,
  excludeTaskId,
  initialDependencyTasks = [],
  memberOptions,
  descriptionInitialValue,
  descriptionEditorKey,
}: TaskFormFieldsProps) => {
  const {
    control,
    register,
    setValue,
    formState: { errors },
  } = useFormContext<ITaskForm>();

  return (
    <div className="grid gap-6 lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.35fr)]">
      <div className="flex min-w-0 flex-col gap-4">
        <Input
          label="Title"
          id="title"
          placeholder="Enter task title"
          {...register('title')}
          error={errors.title?.message}
        />
        <TaskCodeAutocomplete projectId={projectId} />
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
        <div className="grid w-full gap-4 sm:grid-cols-2">
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
              <div className="space-y-3">
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
              </div>
            );
          }}
        />
        <div className="border-border bg-card space-y-4 rounded-2xl border p-4 shadow-sm">
          <div className="space-y-1">
            <h3 className="text-foreground text-sm font-semibold">Dependencies</h3>
            <p className="text-muted-foreground text-sm">
              Link tasks that should be finished before this one moves forward.
            </p>
          </div>
          <TaskDependencyPicker
            projectId={projectId}
            excludeTaskId={excludeTaskId}
            initialSelectedTasks={initialDependencyTasks}
          />
        </div>
      </div>
      <div className="min-w-0 lg:pl-2">
        <TextEditor
          key={descriptionEditorKey}
          initialValue={descriptionInitialValue}
          onChange={(html) => setValue('description', html)}
          label="Description"
          id="description"
          placeholder="Enter task description"
          error={errors.description?.message}
        />
      </div>
    </div>
  );
};
