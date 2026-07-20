import { api } from '../../../shared/services/api';
import type { User } from '@/shared/types/user';
import type { IChangePasswordForm } from '@/features/auth/schemas/change-password.schema';
import type { ISignUpForm } from '@/features/auth/schemas/sign-up.schema';

export const getMe = async () => {
  const response = await api.get('users/me');
  const json = await response.json<User>();
  return json;
};

export const createUser = async (form: ISignUpForm) => {
  const payload = {
    name: form.name,
    email: form.email,
    password: form.password,
  };

  const response = await api.post('users', {
    json: payload,
  });

  const json = await response.json<User>();

  return json;
};

export const changePassword = async (form: IChangePasswordForm) => {
  await api.put('users/me/password', {
    json: form,
  });
};

export const listUsers = async (): Promise<User[]> => {
  const response = await api.get('users');
  return response.json<User[]>();
};
