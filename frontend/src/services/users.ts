import type { User } from '@/types/user';
import { api } from './api';

export const getMe = async () => {
  const response = await api.get('users/me');
  const json = await response.json<User>();
  return json;
};
