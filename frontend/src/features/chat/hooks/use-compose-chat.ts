import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useMemo, useState } from 'react';
import { useMessagesSheet } from '@/features/chat/components/messages-sheet-context';
import { getOrCreateGeneralChat } from '@/features/chat/services/general-chat';
import { generalChatQueryKeys, userListQueryKeys } from '@/shared/services/query-keys';
import { listUsers } from '@/features/auth/services/users';
import { handleError } from '@/shared/utils/handle-error';

export const useComposeChat = () => {
  const queryClient = useQueryClient();
  const { openChat, backToList } = useMessagesSheet();
  const [search, setSearch] = useState('');
  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);

  const { data: users = [] } = useQuery({
    queryKey: userListQueryKeys.all,
    queryFn: listUsers,
  });

  const filteredUsers = useMemo(() => {
    const query = search.toLowerCase();

    return users.filter((user) => user.name.toLowerCase().includes(query) || user.email.toLowerCase().includes(query));
  }, [search, users]);

  const { mutate: startChat, isPending } = useMutation({
    mutationFn: (userIds: string[]) => getOrCreateGeneralChat(userIds),
    onSuccess: (chat) => {
      queryClient.invalidateQueries({ queryKey: generalChatQueryKeys.list });
      openChat(chat.id);
    },
    onError: handleError,
  });

  const toggleUser = (userId: string) => {
    setSelectedUserIds((current) =>
      current.includes(userId) ? current.filter((id) => id !== userId) : [...current, userId],
    );
  };

  const closeCompose = () => {
    setSearch('');
    setSelectedUserIds([]);
    backToList();
  };

  const submit = () => {
    if (selectedUserIds.length === 0) {
      return;
    }

    startChat(selectedUserIds);
  };

  return {
    search,
    setSearch,
    selectedUserIds,
    filteredUsers,
    isPending,
    toggleUser,
    submit,
    closeCompose,
  };
};
