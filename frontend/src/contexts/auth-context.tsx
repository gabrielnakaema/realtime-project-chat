import { LoadingSpinner } from '@/components/loading';
import { tokenService } from '@/services/api';
import { attemptRefreshToken } from '@/services/auth';
import { userQueryKeys } from '@/services/query-keys';
import { getMe } from '@/services/users';
import type { LoginResponse } from '@/types/auth';
import type { User } from '@/types/user';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { createContext, useEffect, useState } from 'react';

type AuthStatus = 'loading' | 'authenticated' | 'unauthenticated' | 'error';

interface AuthContextData {
  authenticate: (loginResponse: LoginResponse) => void;
  isAuthenticated: boolean;
  logout: () => void;
  user?: User;
  authStatus: AuthStatus;
}

export const AuthContext = createContext<AuthContextData>({} as AuthContextData);

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  const [authStatus, setAuthStatus] = useState<AuthStatus>('loading');
  const isAuthenticated = authStatus === 'authenticated';
  const { data } = useQuery({
    queryKey: userQueryKeys.me,
    queryFn: getMe,
    enabled: authStatus === 'authenticated',
  });
  const navigate = useNavigate();

  const initAuth = async () => {
    try {
      const refreshTokenResponse = await attemptRefreshToken();
      if (refreshTokenResponse.access_token) {
        tokenService.setToken(refreshTokenResponse.access_token);
        setAuthStatus('authenticated');
      } else {
        setAuthStatus('unauthenticated');
      }
    } catch (error) {
      setAuthStatus('unauthenticated');
    }
  };

  useEffect(() => {
    initAuth();
  }, []);

  const authenticate = (loginResponse: LoginResponse) => {
    tokenService.setToken(loginResponse.access_token);
    setAuthStatus('authenticated');
  };

  const logout = () => {
    setAuthStatus('unauthenticated');
    tokenService.setToken('');
    navigate({ to: '/' });
  };

  if (authStatus === 'loading') {
    return (
      <main className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 flex items-center justify-center p-4 dark:text-slate-100">
        <LoadingSpinner size="4em" />
      </main>
    );
  }

  return (
    <AuthContext.Provider value={{ authenticate, isAuthenticated, logout, user: data, authStatus }}>
      {children}
    </AuthContext.Provider>
  );
};
