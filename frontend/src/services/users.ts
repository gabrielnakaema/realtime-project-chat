import { api } from './api';
import type { User } from '@/types/user';
import type { ISignUpForm } from '@/schemas/sign-up-schema';

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
