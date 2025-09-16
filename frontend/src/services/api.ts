import ky from 'ky';
import { attemptRefreshToken } from './auth';

export interface ApiError {
  message: string;
  status: number;
  name: string;
}

export const baseApiUrl = import.meta.env.VITE_API_URL;

let isRefreshing = false;
let failedRequests: Array<{
  resolve: (value: unknown) => void;
  reject: (reason?: unknown) => void;
}> = [];
/* let token = '';

export const setToken = (token: string) => {
  token = token;
}; */

const processQueue = (error: unknown, token?: string) => {
  failedRequests.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });

  failedRequests = [];
};

class TokenService {
  private _token = '';

  setToken(token: string) {
    this._token = token;
  }

  get token() {
    return this._token;
  }
}

export const tokenService = new TokenService();

export const api = ky.create({
  prefixUrl: baseApiUrl,
  hooks: {
    beforeRequest: [
      (request) => {
        if (!request.url.includes('auth/login') && !request.url.includes('auth/refresh-token')) {
          if (tokenService.token) {
            request.headers.set('Authorization', `Bearer ${tokenService.token}`);
          }
        }
      },
    ],
    afterResponse: [
      async (request, options, response) => {
        if (response.status !== 401) {
          return response;
        }

        if (response.url.includes('auth/login') || response.url.includes('auth/refresh-token')) {
          return response;
        }

        if ((options as any).isRetry) {
          return response;
        }

        if (isRefreshing) {
          return new Promise((resolve, reject) => {
            failedRequests.push({ resolve, reject });
          })
            .then((token) => {
              request.headers.set('Authorization', `Bearer ${token}`);
              (options as any).isRetry = true;
              return ky(request, options);
            })
            .catch((err) => {
              Promise.reject(err);
            });
        }

        isRefreshing = true;

        try {
          const refreshTokenResponse = await attemptRefreshToken();

          processQueue(null, refreshTokenResponse.access_token);

          request.headers.set('Authorization', `Bearer ${refreshTokenResponse.access_token}`);
          tokenService.setToken(refreshTokenResponse.access_token);
          (options as any).isRetry = true;

          return ky(request, options);
        } catch (error) {
          processQueue(error);
          throw error;
        } finally {
          isRefreshing = false;
        }
      },
    ],
  },
});
