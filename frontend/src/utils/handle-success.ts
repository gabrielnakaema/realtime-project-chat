import { toast } from 'react-toastify';

export const handleSuccess = (message: string) => {
  toast.success(message);
};
