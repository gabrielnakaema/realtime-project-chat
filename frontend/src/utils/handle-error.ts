import { HTTPError } from 'ky';
import { toast } from 'react-toastify';

export interface ApiError {
  message: string;
  status: number;
}

export const handleError = async (error: unknown) => {
  const errorMessage = await getErrorMessage(error);

  notifyError(errorMessage);
};

export const getErrorMessage = async (error: unknown) => {
  if (error instanceof HTTPError) {
    let errorMessage = '';

    try {
      const json = await error.response.json<ApiError>();
      errorMessage = json.message;
    } catch (err) {
      const text = await error.response.text();
      errorMessage = text;
    }

    if (errorMessage === '') {
      errorMessage = error?.message || 'An unknown error occurred';
    }

    return errorMessage;
  }

  if (error instanceof Error) {
    return error.message;
  }

  if (typeof error === 'string') {
    return error;
  }

  return 'An unknown error occurred';
};

export const notifyError = (message: string) => {
  toast.error(message);
};
