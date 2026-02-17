'use client';

import { useForm } from 'react-hook-form';
import { useRouter } from 'next/navigation';

import { zodResolver } from '@hookform/resolvers/zod';
import { Mail, Lock } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { toast } from 'sonner';

import { useAuth } from '@/lib/hooks/use-auth';
import { cn } from '@/lib/utils/cn';
import { extractErrorMessage } from '@/lib/utils/error';
import { registerSchema } from '@/lib/validations/auth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import type { RegisterFormData } from '@/lib/validations/auth';

function getPasswordStrength(password: string): number {
  let score = 0;
  if (password.length >= 8) score++;
  if (/[A-Z]/.test(password)) score++;
  if (/[a-z]/.test(password)) score++;
  if (/\d/.test(password)) score++;
  if (/[^A-Za-z0-9]/.test(password)) score++;
  return score;
}

const STRENGTH_COLORS = [
  'bg-muted',
  'bg-destructive',
  'bg-destructive',
  'bg-warning',
  'bg-success',
  'bg-success',
] as const;

const STRENGTH_LABELS = ['', 'weak', 'weak', 'fair', 'good', 'strong'] as const;

export function RegisterForm() {
  const t = useTranslations('auth');
  const router = useRouter();
  const { register: registerUser, isRegistering } = useAuth();

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: { email: '', password: '', confirmPassword: '' },
  });

  const passwordValue = watch('password');
  const strength = getPasswordStrength(passwordValue);

  const onSubmit = async (data: RegisterFormData) => {
    try {
      await registerUser(data);
      toast.success(t('registerSuccess'));
      router.push('/profile-select');
    } catch (error: unknown) {
      toast.error(extractErrorMessage(error));
    }
  };

  const strengthLabel = STRENGTH_LABELS[strength];

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
      <div>
        <Input
          label={t('password')}
          type="password"
          autoComplete="new-password"
          placeholder="••••••••"
          leftIcon={<Lock className="h-4 w-4" />}
          error={errors.password?.message}
          {...register('password')}
        />
        {passwordValue.length > 0 && (
          <div className="mt-2">
            <div className="flex gap-1">
              {Array.from({ length: 5 }).map((_, i) => (
                <div
                  key={`strength-${i}`}
                  className={cn(
                    'h-1 flex-1 rounded-full transition-colors',
                    i < strength ? STRENGTH_COLORS[strength] : 'bg-muted'
                  )}
                />
              ))}
            </div>
            {strengthLabel && (
              <p className="mt-1 text-xs text-muted-foreground">
                {t('passwordStrength')}: {t(strengthLabel)}
              </p>
            )}
          </div>
        )}
      </div>
      <Input
        label={t('confirmPassword')}
        type="password"
        autoComplete="new-password"
        placeholder="••••••••"
        leftIcon={<Lock className="h-4 w-4" />}
        error={errors.confirmPassword?.message}
        {...register('confirmPassword')}
      />
      <Button
        type="submit"
        isLoading={isRegistering}
        className="mt-2 w-full"
        size="lg"
      >
        {t('register')}
      </Button>
    </form>
  );
}
