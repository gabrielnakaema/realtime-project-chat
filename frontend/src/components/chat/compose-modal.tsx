import { Check, Search, X } from 'lucide-react';
import { getAvatarColorClass, isUserSelected } from './utils';
import { useComposeChat } from '@/hooks/use-compose-chat';
import { cn } from '@/lib/utils';

export const MessagesComposeModal = () => {
  const { search, setSearch, selectedUserIds, filteredUsers, isPending, toggleUser, submit, closeCompose } =
    useComposeChat();

  return (
    <div
      className="absolute inset-0 z-10 flex items-start justify-center bg-black/20 pt-14 backdrop-blur-[2px]"
      onClick={(event) => {
        if (event.target === event.currentTarget) {
          closeCompose();
        }
      }}
    >
      <div className="mx-3 w-full overflow-hidden rounded-2xl bg-white shadow-2xl dark:bg-slate-800">
        <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3 dark:border-slate-700">
          <div>
            <span className="text-sm font-semibold text-slate-900 dark:text-slate-100">New chat</span>
            <p className="text-xs text-slate-400">Select one or more teammates</p>
          </div>
          <button
            onClick={closeCompose}
            className="flex h-6 w-6 items-center justify-center rounded-full text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600 dark:hover:bg-slate-700"
          >
            <X className="h-3.5 w-3.5" />
          </button>
        </div>

        <div className="border-b border-slate-100 px-4 py-2 dark:border-slate-700">
          <div className="flex items-center gap-2">
            <Search className="h-3.5 w-3.5 shrink-0 text-slate-400" />
            <input
              autoFocus
              type="text"
              placeholder="Search people..."
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              className="w-full bg-transparent py-1 text-sm text-slate-900 placeholder-slate-400 outline-none dark:text-slate-100"
            />
          </div>
          {selectedUserIds.length > 0 && (
            <div className="mt-2 flex items-center justify-between">
              <p className="text-xs text-slate-500 dark:text-slate-400">{selectedUserIds.length} selected</p>
              <button
                type="button"
                onClick={submit}
                disabled={isPending}
                className="rounded-lg bg-blue-600 px-3 py-1 text-xs font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
              >
                Start chat
              </button>
            </div>
          )}
        </div>

        <ul className="max-h-64 overflow-y-auto py-1">
          {filteredUsers.length === 0 ? (
            <li className="px-4 py-5 text-center text-sm text-slate-400">
              {search ? 'No people found' : 'No users available'}
            </li>
          ) : (
            filteredUsers.map((user) => (
              <li key={user.id}>
                <button
                  type="button"
                  disabled={isPending}
                  onClick={() => toggleUser(user.id)}
                  className="flex w-full items-center gap-3 px-4 py-2 text-left transition-colors hover:bg-slate-50 disabled:opacity-50 dark:hover:bg-slate-700/50"
                >
                  <div
                    className={cn(
                      'flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-semibold text-white',
                      getAvatarColorClass(user.name),
                    )}
                  >
                    {user.name.charAt(0).toUpperCase()}
                  </div>
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium text-slate-900 dark:text-slate-100">{user.name}</p>
                    <p className="truncate text-xs text-slate-400">{user.email}</p>
                  </div>
                  {isUserSelected(selectedUserIds, user.id) && (
                    <div className="ml-auto flex h-6 w-6 items-center justify-center rounded-full bg-blue-600 text-white">
                      <Check className="h-3.5 w-3.5" />
                    </div>
                  )}
                </button>
              </li>
            ))
          )}
        </ul>
      </div>
    </div>
  );
};
