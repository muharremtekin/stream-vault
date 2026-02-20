'use client';

import { Check, Star } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { formatCurrency } from '@/lib/utils/format';
import { cn } from '@/lib/utils/cn';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import type { Plan } from '@/lib/types/subscription';

const TIER_ORDER: Record<string, number> = { Basic: 1, Standard: 2, Premium: 3 };

function getButtonLabel(
  isCurrent: boolean,
  isUpgrade: boolean,
  hasCurrentTier: boolean,
  t: (key: string) => string
): string {
  if (isCurrent) return t('currentPlan');
  if (isUpgrade) return t('upgrade');
  if (hasCurrentTier) return t('downgrade');
  return t('select');
}

interface PlanCardsProps {
  plans: Plan[];
  currentTier: string | null;
  onSelectPlan: (planId: string) => void;
  isLoading: boolean;
}

export function PlanCards({ plans, currentTier, onSelectPlan, isLoading }: PlanCardsProps) {
  const t = useTranslations('subscription.plans');
  const tStatus = useTranslations('subscription.status');

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
        {['skeleton-plan-1', 'skeleton-plan-2', 'skeleton-plan-3'].map((id) => (
          <div key={id} className="rounded-lg border border-border bg-card p-6">
            <Skeleton variant="text" className="mb-4 h-6 w-24" />
            <Skeleton variant="text" className="mb-6 h-10 w-32" />
            <div className="space-y-3">
              <Skeleton variant="text" className="h-4 w-full" />
              <Skeleton variant="text" className="h-4 w-full" />
              <Skeleton variant="text" className="h-4 w-3/4" />
            </div>
            <Skeleton variant="text" className="mt-6 h-10 w-full" />
          </div>
        ))}
      </div>
    );
  }

  const sortedPlans = [...plans]
    .filter((p) => p.isActive)
    .sort((a, b) => (TIER_ORDER[a.tier] ?? 0) - (TIER_ORDER[b.tier] ?? 0));

  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
      {sortedPlans.map((plan) => {
        const isCurrent = currentTier === plan.tier;
        const currentOrder = currentTier ? (TIER_ORDER[currentTier] ?? 0) : 0;
        const planOrder = TIER_ORDER[plan.tier] ?? 0;
        const isUpgrade = planOrder > currentOrder;
        const isPopular = plan.tier === 'Standard';
        const features = plan.features
          ? plan.features.split(',').map((f) => f.trim()).filter(Boolean)
          : [];

        return (
          <div
            key={plan.id}
            className={cn(
              'relative flex flex-col rounded-lg border bg-card p-6 transition-colors',
              isCurrent
                ? 'border-primary ring-1 ring-primary'
                : 'border-border hover:border-muted-foreground/50'
            )}
          >
            {isPopular && (
              <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                <Badge variant="warning" size="sm">
                  <Star className="mr-1 h-3 w-3" />
                  {t('popular')}
                </Badge>
              </div>
            )}

            <div className="mb-4">
              <h3 className="text-lg font-semibold text-foreground">{plan.name}</h3>
              {isCurrent && (
                <Badge variant="success" size="sm" className="mt-1">
                  {t('currentPlan')}
                </Badge>
              )}
            </div>

            <div className="mb-6">
              <span className="text-3xl font-bold text-foreground">
                {formatCurrency(plan.priceMonthly)}
              </span>
              <span className="text-sm text-muted-foreground">
                {tStatus('perMonth')}
              </span>
            </div>

            <div className="mb-4 flex items-center gap-4 text-sm text-muted-foreground">
              <span>{plan.maxQuality}</span>
              <span className="text-border">|</span>
              <span>
                {tStatus('screens', { count: plan.maxScreens })}
              </span>
            </div>

            <ul className="mb-6 flex-1 space-y-2">
              {features.map((feature) => (
                <li key={feature} className="flex items-center gap-2 text-sm text-foreground">
                  <Check className="h-4 w-4 shrink-0 text-primary" />
                  {feature}
                </li>
              ))}
            </ul>

            <Button
              variant={isCurrent || !isUpgrade ? 'secondary' : 'primary'}
              className="w-full"
              disabled={isCurrent}
              onClick={() => onSelectPlan(plan.id)}
            >
              {getButtonLabel(isCurrent, isUpgrade, !!currentTier, t)}
            </Button>
          </div>
        );
      })}
    </div>
  );
}
