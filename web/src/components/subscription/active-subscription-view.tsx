'use client';

import { useTranslations } from 'next-intl';

import { SubscriptionStatus } from '@/components/subscription/subscription-status';
import { PlanCards } from '@/components/subscription/plan-cards';
import { ChangePlanModal } from '@/components/subscription/change-plan-modal';
import { CancelModal } from '@/components/subscription/cancel-modal';
import { InvoiceTable } from '@/components/subscription/invoice-table';
import type { Plan, Subscription, Invoice } from '@/lib/types/subscription';

interface ActiveSubscriptionViewProps {
  subscription: Subscription;
  plans: Plan[];
  invoices: Invoice[];
  currentTier: string | null;
  isLoadingPlans: boolean;
  isChangingPlan: boolean;
  isCancelling: boolean;
  changePlan: Plan | undefined;
  isChangePlanOpen: boolean;
  isCancelOpen: boolean;
  onSelectPlan: (planId: string) => void;
  onChangePlanConfirm: (cardNumber: string) => void;
  onChangePlanClose: () => void;
  onCancelOpen: () => void;
  onCancelClose: () => void;
  onCancelConfirm: () => void;
}

export function ActiveSubscriptionView({
  subscription,
  plans,
  invoices,
  currentTier,
  isLoadingPlans,
  isChangingPlan,
  isCancelling,
  changePlan,
  isChangePlanOpen,
  isCancelOpen,
  onSelectPlan,
  onChangePlanConfirm,
  onChangePlanClose,
  onCancelOpen,
  onCancelClose,
  onCancelConfirm,
}: ActiveSubscriptionViewProps) {
  const t = useTranslations('subscription');
  const currentPlan = subscription.plan;

  return (
    <>
      <SubscriptionStatus
        subscription={subscription}
        onChangePlan={() => {
          document.getElementById('plan-cards-section')?.scrollIntoView({ behavior: 'smooth' });
        }}
        onCancel={onCancelOpen}
        isCancelling={isCancelling}
      />

      <div id="plan-cards-section">
        <h3 className="mb-4 text-lg font-semibold text-foreground">
          {t('plans.changePlan')}
        </h3>
        <PlanCards
          plans={plans}
          currentTier={currentTier}
          onSelectPlan={onSelectPlan}
          isLoading={isLoadingPlans}
        />
      </div>

      <InvoiceTable invoices={invoices} isLoading={false} />

      {changePlan && (
        <ChangePlanModal
          isOpen={isChangePlanOpen}
          onClose={onChangePlanClose}
          currentPlan={currentPlan}
          newPlan={changePlan}
          onConfirm={onChangePlanConfirm}
          isChanging={isChangingPlan}
        />
      )}

      <CancelModal
        isOpen={isCancelOpen}
        onClose={onCancelClose}
        onConfirm={onCancelConfirm}
        isCancelling={isCancelling}
        periodEnd={subscription.periodEnd}
      />
    </>
  );
}
