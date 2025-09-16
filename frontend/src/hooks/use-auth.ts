import { AuthContext } from '@/contexts/auth-context';
import { useContext } from 'react';

export const useAuth = () => {
  const data = useContext(AuthContext);

  if (!data) {
    throw new Error('useAuth must be used within an AuthProvider');
  }

  return data;
};
