/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import { useContext } from 'react';
import { AuthContext } from '@/contexts/auth-context';

export const useAuth = () => {
  const data = useContext(AuthContext);

  if (!data) {
    throw new Error('useAuth must be used within an AuthProvider');
  }

  return data;
};
