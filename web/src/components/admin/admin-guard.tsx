'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';

import { useAuthStore } from '@/lib/stores/auth-store';
import { SUBSCRIPTION_TIERS } from '@/lib/types/common';

interface AdminGuardProps {
  children: React.ReactNode;
}

export function AdminGuard({ children }: AdminGuardProps) {
  const t = useTranslations('admin');
  const router = useRouter();
  const user = useAuthStore((s) => s.user);
  const accessToken = useAuthStore((s) => s.accessToken);

  const isAdmin = user?.role === SUBSCRIPTION_TIERS.Admin;
  const isHydrating = accessToken !== null && user === null;

  useEffect(() => {
    if (!isHydrating && !isAdmin) {
      router.replace('/browse');
    }
  }, [isAdmin, isHydrating, router]);

  if (isHydrating) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <p className="text-muted-foreground">{t('redirecting')}</p>
      </div>
    );
  }

  if (!isAdmin) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <p className="text-muted-foreground">{t('accessDenied')}</p>
      </div>
    );
  }

  return <>{children}</>;
}
