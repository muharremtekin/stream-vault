import { toast } from 'sonner';

export { toast };

export function showSuccessToast(message: string): void {
  toast.success(message);
}

export function showErrorToast(message: string): void {
  toast.error(message);
}

export function showInfoToast(message: string): void {
  toast.info(message);
}
