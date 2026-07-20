import { Check, Search, X } from 'lucide-react';
import { getAvatarColorClass, isUserSelected } from './utils';
import { useComposeChat } from '@/features/chat/hooks/use-compose-chat';
import { cn } from '@/lib/utils';

export const ComposeModal = () => {
  const { search, setSearch, selectedUserIds, filteredUsers, isPending, toggleUser, submit, closeCompose } =
    useComposeChat();

  return (
    <div
      className="bg-overlay/20 absolute inset-0 z-10 flex items-start justify-center pt-14 backdrop-blur-[2px]"
      onClick={(event) => {
        if (event.target === event.currentTarget) {
          closeCompose();
        }
      }}
    >
      <div className="bg-card mx-3 w-full overflow-hidden rounded-2xl shadow-2xl">
        <div className="border-border flex items-center justify-between border-b px-4 py-3">
          <div>
            <span className="text-foreground text-sm font-semibold">New chat</span>
            <p className="text-muted-foreground text-xs">Select one or more teammates</p>
          </div>
          <button
            onClick={closeCompose}
            className="text-muted-foreground hover:bg-muted hover:text-muted-foreground flex h-6 w-6 items-center justify-center rounded-full transition-colors"
          >
            <X className="h-3.5 w-3.5" />
          </button>
        </div>

        <div className="border-border border-b px-4 py-2">
          <div className="flex items-center gap-2">
            <Search className="text-muted-foreground h-3.5 w-3.5 shrink-0" />
            <input
              autoFocus
              type="text"
              placeholder="Search people..."
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              className="text-foreground placeholder-muted-foreground w-full bg-transparent py-1 text-sm outline-none"
            />
          </div>
          {selectedUserIds.length > 0 && (
            <div className="mt-2 flex items-center justify-between">
              <p className="text-muted-foreground text-xs">{selectedUserIds.length} selected</p>
              <button
                type="button"
                onClick={submit}
                disabled={isPending}
                className="bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg px-3 py-1 text-xs font-medium transition-colors disabled:opacity-50"
              >
                Start chat
              </button>
            </div>
          )}
        </div>

        <ul className="max-h-64 overflow-y-auto py-1">
          {filteredUsers.length === 0 ? (
            <li className="text-muted-foreground px-4 py-5 text-center text-sm">
              {search ? 'No people found' : 'No users available'}
            </li>
          ) : (
            filteredUsers.map((user) => (
              <li key={user.id}>
                <button
                  type="button"
                  disabled={isPending}
                  onClick={() => toggleUser(user.id)}
                  className="hover:bg-muted flex w-full items-center gap-3 px-4 py-2 text-left transition-colors disabled:opacity-50"
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
                    <p className="text-foreground truncate text-sm font-medium">{user.name}</p>
                    <p className="text-muted-foreground truncate text-xs">{user.email}</p>
                  </div>
                  {isUserSelected(selectedUserIds, user.id) && (
                    <div className="bg-primary text-primary-foreground ml-auto flex h-6 w-6 items-center justify-center rounded-full">
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
