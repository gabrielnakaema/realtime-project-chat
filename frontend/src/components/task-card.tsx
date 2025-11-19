import { Calendar } from 'lucide-react';
import { useState } from 'react';
import { Avatar } from './avatar';
import { TaskDetails } from './task-details';
import { TaskPriorityBadge } from './task-priority-badge';
import { EditTask } from './edit-task';
import type { Task } from '@/types/task';

interface TaskCardProps {
  task: Task;
  onDragStart: () => void;
}

export const TaskCard = ({ task, onDragStart }: TaskCardProps) => {
  const [open, setOpen] = useState(false);
  const [openEdit, setOpenEdit] = useState(false);

  console.log('openEdit', openEdit);
  console.log('open', open);
  return (
    <>
      <div
        className="cursor-pointer rounded-lg border border-slate-200 bg-white p-4 transition-shadow hover:shadow-md dark:border-slate-700 dark:bg-slate-900"
        draggable
        onDragStart={onDragStart}
        onClick={() => setOpen(true)}
      >
        <div className="pb-3">
          <div className="flex items-start justify-between">
            <h4 className="text-sm leading-tight font-medium text-slate-900 dark:text-slate-100">{task.title}</h4>
            <TaskPriorityBadge priority={task.priority} />
          </div>
          {task.description && (
            <p className="mt-2 line-clamp-2 text-xs text-slate-600 dark:text-slate-400">{task.description}</p>
          )}

          {!!task.tags?.length && (
            <div className="flex w-full flex-wrap items-center gap-2 pt-2">
              {task.tags.map((tag) => (
                <div
                  key={tag}
                  className="w-fit rounded-sm border border-slate-200 px-2 py-0.5 text-xs font-medium text-slate-500 dark:border-slate-700 dark:text-slate-400"
                >
                  {tag}
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">{task.responsible && <Avatar name={task.responsible.name} />}</div>

          {task.due_date && (
            <div className="flex items-center gap-1 text-xs text-slate-500 dark:text-slate-400">
              <Calendar className="h-3 w-3" />
              <span>{new Date(task.due_date).toLocaleDateString()}</span>
            </div>
          )}
        </div>
      </div>
      <TaskDetails
        taskId={task.id}
        open={open}
        onOpenChange={setOpen}
        onEdit={() => {
          setOpenEdit(true);
          setOpen(false);
        }}
      />
      <EditTask taskId={task.id} open={openEdit} onOpenChange={setOpenEdit} />
    </>
  );
};
