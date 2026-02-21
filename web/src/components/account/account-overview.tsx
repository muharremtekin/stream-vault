'use client';

import Link from 'next/link';
import {
  Users,
  CreditCard,
  Bell,
  Mail,
  ChevronRight,
  Shield,
} from 'lucide-react';
import { useTranslations } from 'next-intl';

import { useAuthStore } from '@/lib/stores/auth-store';
import { useSubscription } from '@/lib/hooks/use-subscription';
import { cn } from '@/lib/utils/cn';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';

const NAV_SECTIONS = [
  {
    key: 'profiles' as const,
    href: '/account/profiles',
    icon: Users,
  },
  {
    key: 'subscription' as const,
    href: '/account/subscription',
    icon: CreditCard,
  },
  {
    key: 'notifications' as const,
    href: '/account/notifications',
    icon: Bell,
  },
] as const;

export function AccountOverview() {
  const t = useTranslations('account');
  const user = useAuthStore((s) => s.user);
  const subscriptionTier = useAuthStore((s) => s.subscriptionTier);
  const { subscription, isLoadingSubscription } = useSubscription();

  return (
    <div className="space-y-8">
      {/* User Info Card */}
      <div className="rounded-lg border border-border bg-card p-6">
        <div className="flex items-start gap-4">
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-primary/10">
            <Mail className="h-7 w-7 text-primary" />
          </div>
          <div className="flex-1">
            <p className="text-lg font-semibold text-foreground">
              {user?.email}
            </p>
            <div className="mt-2 flex items-center gap-2">
              <Shield className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm text-muted-foreground">
                {t('subscriptionTier')}:
              </span>
              <Badge variant={subscriptionTier === 'Free' ? 'default' : 'success'}>
                {subscriptionTier ?? 'Free'}
              </Badge>
            </div>
          </div>
        </div>

        {/* Subscription Summary */}
        <div className="mt-6 border-t border-border pt-4">
          {isLoadingSubscription ? (
            <Skeleton variant="text" className="h-5 w-48" />
          ) : subscription ? (
            <div className="flex items-center justify-between">
              <div>
                <span className="text-sm text-muted-foreground">
                  {t('currentPlan')}:
                </span>
                <span className="ml-2 font-medium text-foreground">
                  {subscription.plan.name}
                </span>
              </div>
              <Link
                href="/account/subscription"
                className="text-sm text-primary hover:underline focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
              >
                {t('sections.subscription')}
              </Link>
            </div>
          ) : (
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">
                {t('noSubscription')}
              </span>
              <Link
                href="/account/subscription"
                className="text-sm font-medium text-primary hover:underline focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
              >
                {t('subscribeCta')}
              </Link>
            </div>
          )}
        </div>
      </div>

      {/* Navigation Sections */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        {NAV_SECTIONS.map((section) => {
          const Icon = section.icon;
          return (
            <Link
              key={section.key}
              href={section.href}
              className={cn(
                'group flex items-center gap-4 rounded-lg border border-border bg-card p-5',
                'transition-colors hover:border-primary/50 hover:bg-card/80',
                'focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none'
              )}
            >
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-muted">
                <Icon className="h-5 w-5 text-muted-foreground group-hover:text-primary" />
              </div>
              <div className="flex-1">
                <p className="font-medium text-foreground">
                  {t(`sections.${section.key}`)}
                </p>
                <p className="text-xs text-muted-foreground">
                  {t(`sections.${section.key}Desc`)}
                </p>
              </div>
              <ChevronRight className="h-5 w-5 text-muted-foreground group-hover:text-foreground" />
            </Link>
          );
        })}
      </div>
    </div>
  );
}
