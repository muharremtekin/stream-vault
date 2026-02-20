'use client';

import { useEffect } from 'react';

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Mail, Smartphone, MonitorPlay } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { getPreferences, updatePreferences } from '@/lib/api/notification';
import {
  notificationPreferencesSchema,
} from '@/lib/validations/notification-prefs-schema';
import { showSuccessToast, showErrorToast } from '@/components/ui/toast';
import { extractErrorMessage } from '@/lib/utils/error';
import { cn } from '@/lib/utils/cn';
import type { NotificationPreferencesFormData } from '@/lib/validations/notification-prefs-schema';
import type { LucideIcon } from 'lucide-react';

interface ChannelRowProps {
  icon: LucideIcon;
  label: string;
  description: string;
  isEnabled: boolean;
  onToggle: (enabled: boolean) => void;
}

function ChannelRow({ icon: Icon, label, description, isEnabled, onToggle }: ChannelRowProps) {
  return (
    <div className="flex items-center justify-between gap-4 py-4">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-muted">
          <Icon className="h-5 w-5 text-muted-foreground" />
        </div>
        <div>
          <p className="text-sm font-medium text-foreground">{label}</p>
          <p className="text-xs text-muted-foreground">{description}</p>
        </div>
      </div>
      <button
        type="button"
        role="switch"
        aria-checked={isEnabled}
        onClick={() => onToggle(!isEnabled)}
        className={cn(
          'relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background focus-visible:outline-none',
          isEnabled ? 'bg-primary' : 'bg-muted'
        )}
      >
        <span
          className={cn(
            'pointer-events-none inline-block h-5 w-5 rounded-full bg-foreground shadow-lg transition-transform',
            isEnabled ? 'translate-x-5' : 'translate-x-0.5'
          )}
          style={{ marginTop: '2px' }}
        />
      </button>
    </div>
  );
}

export function NotificationPrefs() {
  const t = useTranslations('notifications.preferences');
  const tApi = useTranslations('api');
  const queryClient = useQueryClient();

  const { data: preferences, isLoading } = useQuery({
    queryKey: ['notifications', 'preferences'],
    queryFn: getPreferences,
    staleTime: 60_000,
  });

  const { watch, setValue, handleSubmit, reset } = useForm<NotificationPreferencesFormData>({
    resolver: zodResolver(notificationPreferencesSchema),
    defaultValues: {
      email: { enabled: true },
      push: { enabled: true },
      inApp: { enabled: true },
    },
  });

  useEffect(() => {
    if (preferences) {
      reset({
        email: { enabled: preferences.email.enabled, categories: preferences.email.categories },
        push: { enabled: preferences.push.enabled, categories: preferences.push.categories },
        inApp: { enabled: preferences.inApp.enabled, categories: preferences.inApp.categories },
      });
    }
  }, [preferences, reset]);

  const mutation = useMutation({
    mutationFn: updatePreferences,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications', 'preferences'] });
      showSuccessToast(tApi('preferencesUpdated'));
    },
    onError: (error) => showErrorToast(extractErrorMessage(error)),
  });

  const onSubmit = (formData: NotificationPreferencesFormData) => {
    mutation.mutate(formData);
  };

  const emailEnabled = watch('email.enabled');
  const pushEnabled = watch('push.enabled');
  const inAppEnabled = watch('inApp.enabled');

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="h-6 w-48 animate-pulse rounded bg-muted" />
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={`pref-skeleton-${i}`} className="flex items-center justify-between py-4">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 animate-pulse rounded-full bg-muted" />
              <div className="space-y-1">
                <div className="h-4 w-32 animate-pulse rounded bg-muted" />
                <div className="h-3 w-48 animate-pulse rounded bg-muted" />
              </div>
            </div>
            <div className="h-6 w-11 animate-pulse rounded-full bg-muted" />
          </div>
        ))}
      </div>
    );
  }

  return (
    <div>
      <h2 className="text-lg font-semibold text-foreground">{t('title')}</h2>
      <p className="mt-1 text-sm text-muted-foreground">{t('description')}</p>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-4">
        <div className="divide-y divide-border">
          <ChannelRow
            icon={Mail}
            label={t('email')}
            description={t('emailDescription')}
            isEnabled={emailEnabled}
            onToggle={(v) => setValue('email.enabled', v, { shouldDirty: true })}
          />
          <ChannelRow
            icon={Smartphone}
            label={t('push')}
            description={t('pushDescription')}
            isEnabled={pushEnabled}
            onToggle={(v) => setValue('push.enabled', v, { shouldDirty: true })}
          />
          <ChannelRow
            icon={MonitorPlay}
            label={t('inApp')}
            description={t('inAppDescription')}
            isEnabled={inAppEnabled}
            onToggle={(v) => setValue('inApp.enabled', v, { shouldDirty: true })}
          />
        </div>

        <button
          type="submit"
          disabled={mutation.isPending}
          className="mt-6 rounded-md bg-primary px-6 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none disabled:opacity-50"
        >
          {mutation.isPending ? t('saving') : t('save')}
        </button>
      </form>
    </div>
  );
}
