'use client';

import { useForm } from 'react-hook-form';
import { useRouter } from 'next/navigation';

import { zodResolver } from '@hookform/resolvers/zod';
import { Mail, Lock } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { toast } from 'sonner';

import { useAuth } from '@/lib/hooks/use-auth';
import { extractErrorMessage } from '@/lib/utils/error';
import { loginSchema } from '@/lib/validations/auth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import type { LoginFormData } from '@/lib/validations/auth';

export function LoginForm() {
  const t = useTranslations('auth');
  const router = useRouter();
  const { login, isLoggingIn } = useAuth();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: '', password: '' },
  });

  const onSubmit = async (data: LoginFormData) => {
    try {
      await login(data);
      toast.success(t('loginSuccess'));
      router.push('/profile-select');
    } catch (error: unknown) {
      toast.error(extractErrorMessage(error));
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} noValidate className="space-y-4">
      <Input
        label={t('email')}
        type="email"
        autoComplete="email"
        placeholder="name@example.com"
        leftIcon={<Mail className="h-4 w-4" />}
        error={errors.email?.message}
        {...register('email')}
      />
      <Input
        label={t('password')}
        type="password"
        autoComplete="current-password"
        placeholder="••••••••"
        leftIcon={<Lock className="h-4 w-4" />}
        error={errors.password?.message}
        {...register('password')}
      />
      <Button
        type="submit"
        isLoading={isLoggingIn}
        className="mt-2 w-full"
        size="lg"
      >
        {t('login')}
      </Button>
    </form>
  );
}
