import type { Chat } from '@/features/chat/types/chat';
import type { User } from '@/shared/types/user';
import { formatMonthDay } from '@/shared/utils/date';

export const AVATAR_COLORS = [
  'bg-violet-500',
  'bg-blue-500',
  'bg-emerald-500',
  'bg-rose-500',
  'bg-amber-500',
  'bg-indigo-500',
  'bg-teal-500',
  'bg-pink-500',
];

export const getAvatarColorClass = (name: string) => {
  let hash = 0;
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash);
  return AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length];
};

export const formatTime = (ts: string) => new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

export const formatDate = (ts: string) => formatMonthDay(ts);

export const getOtherMembers = (chat: Chat, currentUserId?: string): User[] => {
  return chat.members
    .map((member) => member.user)
    .filter((member): member is User => member !== null)
    .filter((member) => member.id !== currentUserId);
};

export const getChatTitle = (chat: Chat, currentUserId?: string) => {
  const members = getOtherMembers(chat, currentUserId);
  if (members.length === 0) return 'You';
  return members.map((member) => member.name).join(', ');
};

export const getChatSubtitle = (chat: Chat, currentUserId?: string) => {
  const members = getOtherMembers(chat, currentUserId);
  if (members.length <= 1) return 'General chat';
  return `${members.length} members`;
};

export const getChatAvatarSeed = (chat: Chat, currentUserId?: string) => {
  const members = getOtherMembers(chat, currentUserId);
  return members[0]?.name ?? 'Chat';
};

export const isUserSelected = (selectedUserIds: string[], userId: string) => {
  return selectedUserIds.includes(userId);
};
