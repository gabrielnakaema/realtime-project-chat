import type { ILoginForm } from '@/schemas/login-schema';
import type { LoginResponse } from '@/types/auth';
import { api } from './api';

export const login = async (form: ILoginForm) => {
  const payload = {
    email: form.email,
    password: form.password,
  };

  const response = await api.post('auth/login', {
    json: payload,
    credentials: 'include',
  });

  const json = await response.json<LoginResponse>();

  return json;
};

export const attemptRefreshToken = async () => {
  const response = await api.post('auth/refresh-token', {
    credentials: 'include',
  });

  const json = await response.json<LoginResponse>();

  return json;
};

export const attemptLogout = async () => {
  await api.post('auth/logout', {
    credentials: 'include',
  });
};
