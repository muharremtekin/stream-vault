'use client';

import {
  CreditCard,
  Calendar,
  Monitor,
  Tv,
  CheckCircle,
  AlertTriangle,
} from 'lucide-react';
import { useTranslations } from 'next-intl';

import { formatCurrency, formatDate } from '@/lib/utils/format';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import type { Subscription } from '@/lib/types/subscription';

const STATUS_BADGE: Record<string, { variant: 'success' | 'warning' | 'error' | 'info'; key: string }> = {
  Active: { variant: 'success', key: 'active' },
  Cancelled: { variant: 'warning', key: 'cancelled' },
  Expired: { variant: 'error', key: 'expired' },
  PendingPayment: { variant: 'info', key: 'pendingPayment' },
  Failed: { variant: 'error', key: 'failed' },
};

interface SubscriptionStatusProps {
  subscription: Subscription;
  onChangePlan: () => void;
  onCancel: () => void;
  isCancelling: boolean;
}

export function SubscriptionStatus({
  subscription,
  onChangePlan,
  onCancel,
  isCancelling,
}: SubscriptionStatusProps) {
  const t = useTranslations('subscription.status');
  const tCancel = useTranslations('subscription.cancel');
  const tPlans = useTranslations('subscription.plans');
  const { plan, status, periodEnd, cancelledAt } = subscription;
  const badge = STATUS_BADGE[status] ?? { variant: 'info' as const, key: status.toLowerCase() };
  const features = plan.features ? plan.features.split(',').map((f) => f.trim()).filter(Boolean) : [];
  const isCancelled = status === 'Cancelled';

  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <div className="flex items-start justify-between">
        <div>
          <h2 className="text-xl font-semibold text-foreground">{t('title')}</h2>
          <div className="mt-1 flex items-center gap-2">
            <span className="text-2xl font-bold text-foreground">{plan.name}</span>
            <Badge variant={badge.variant}>{t(badge.key)}</Badge>
          </div>
        </div>
        <div className="text-right">
          <span className="text-3xl font-bold text-foreground">
            {formatCurrency(plan.priceMonthly)}
          </span>
          <span className="text-sm text-muted-foreground">{t('perMonth')}</span>
        </div>
      </div>

      {isCancelled && cancelledAt && (
        <div className="mt-4 flex items-center gap-2 rounded-md bg-warning/10 px-4 py-3 text-sm text-warning">
          <AlertTriangle className="h-4 w-4 shrink-0" />
          <span>{t('endsOn', { date: formatDate(periodEnd) })}</span>
        </div>
      )}

      <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <div className="flex items-center gap-3">
          <Tv className="h-5 w-5 text-muted-foreground" />
          <div>
            <p className="text-xs text-muted-foreground">{t('maxQuality')}</p>
            <p className="text-sm font-medium text-foreground">{plan.maxQuality}</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Monitor className="h-5 w-5 text-muted-foreground" />
          <div>
            <p className="text-xs text-muted-foreground">{t('maxScreens')}</p>
            <p className="text-sm font-medium text-foreground">
              {t('screens', { count: plan.maxScreens })}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Calendar className="h-5 w-5 text-muted-foreground" />
          <div>
            <p className="text-xs text-muted-foreground">{t('renewalDate')}</p>
            <p className="text-sm font-medium text-foreground">{formatDate(periodEnd)}</p>
          </div>
        </div>
      </div>

      {features.length > 0 && (
        <div className="mt-6">
          <p className="mb-2 text-xs font-medium uppercase text-muted-foreground">
            {t('features')}
          </p>
          <ul className="space-y-1.5">
            {features.map((feature) => (
              <li key={feature} className="flex items-center gap-2 text-sm text-foreground">
                <CheckCircle className="h-4 w-4 text-primary" />
                {feature}
              </li>
            ))}
          </ul>
        </div>
      )}

      <div className="mt-6 flex gap-3">
        {!isCancelled && (
          <>
            <Button variant="secondary" onClick={onChangePlan}>
              <CreditCard className="mr-2 h-4 w-4" />
              {tPlans('changePlan')}
            </Button>
            <Button
              variant="danger"
              onClick={onCancel}
              isLoading={isCancelling}
            >
              {tCancel('title')}
            </Button>
          </>
        )}
      </div>
    </div>
  );
}
