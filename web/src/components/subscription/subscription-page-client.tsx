'use client';

import { useState, useCallback } from 'react';

import { useTranslations } from 'next-intl';
import { toast } from 'sonner';

import { useSubscription } from '@/lib/hooks/use-subscription';
import { useAuthStore } from '@/lib/stores/auth-store';
import { extractErrorMessage } from '@/lib/utils/error';
import { PlanCards } from '@/components/subscription/plan-cards';
import { PaymentForm } from '@/components/subscription/payment-form';
import { ActiveSubscriptionView } from '@/components/subscription/active-subscription-view';

export function SubscriptionPageClient() {
  const t = useTranslations('subscription');
  const tApi = useTranslations('api');

  const {
    plans,
    subscription,
    invoices,
    isLoadingPlans,
    isLoadingSubscription,
    subscribe: subscribeMutate,
    changePlan: changePlanMutate,
    cancel: cancelMutate,
    isSubscribing,
    isChangingPlan,
    isCancelling,
  } = useSubscription();

  const subscriptionTier = useAuthStore((s) => s.subscriptionTier);

  const [selectedPlanId, setSelectedPlanId] = useState<string | null>(null);
  const [showPayment, setShowPayment] = useState(false);
  const [showChangePlan, setShowChangePlan] = useState(false);
  const [showCancel, setShowCancel] = useState(false);
  const [changePlanId, setChangePlanId] = useState<string | null>(null);

  const hasSubscription = !!subscription && subscription.status !== 'Expired';
  const selectedPlan = plans.find((p) => p.id === selectedPlanId);
  const changePlanData = plans.find((p) => p.id === changePlanId);

  const handleSelectPlan = useCallback(
    (planId: string) => {
      if (hasSubscription) {
        setChangePlanId(planId);
        setShowChangePlan(true);
      } else {
        setSelectedPlanId(planId);
        setShowPayment(true);
      }
    },
    [hasSubscription]
  );

  const handleSubscribe = useCallback(
    async (cardNumber: string) => {
      if (!selectedPlanId) return;
      try {
        await subscribeMutate({ planId: selectedPlanId, cardNumber });
        toast.success(tApi('subscriptionCreated'));
        setShowPayment(false);
        setSelectedPlanId(null);
      } catch (error: unknown) {
        toast.error(extractErrorMessage(error));
      }
    },
    [selectedPlanId, subscribeMutate, tApi]
  );

  const handleChangePlanConfirm = useCallback(
    async (cardNumber: string) => {
      if (!changePlanId) return;
      try {
        await changePlanMutate({ newPlanId: changePlanId, cardNumber });
        toast.success(tApi('planChanged'));
        setShowChangePlan(false);
        setChangePlanId(null);
      } catch (error: unknown) {
        toast.error(extractErrorMessage(error));
      }
    },
    [changePlanId, changePlanMutate, tApi]
  );

  const handleCancel = useCallback(async () => {
    try {
      await cancelMutate();
      toast.success(tApi('subscriptionCancelled'));
      setShowCancel(false);
    } catch (error: unknown) {
      toast.error(extractErrorMessage(error));
    }
  }, [cancelMutate, tApi]);

  if (isLoadingSubscription) {
    return (
      <div className="space-y-8">
        <PlanCards plans={[]} currentTier={null} onSelectPlan={() => {}} isLoading />
      </div>
    );
  }

  if (hasSubscription && subscription) {
    return (
      <div className="space-y-8">
        <ActiveSubscriptionView
          subscription={subscription}
          plans={plans}
          invoices={invoices}
          currentTier={subscriptionTier}
          isLoadingPlans={isLoadingPlans}
          isChangingPlan={isChangingPlan}
          isCancelling={isCancelling}
          changePlan={changePlanData}
          isChangePlanOpen={showChangePlan}
          isCancelOpen={showCancel}
          onSelectPlan={handleSelectPlan}
          onChangePlanConfirm={handleChangePlanConfirm}
          onChangePlanClose={() => { setShowChangePlan(false); setChangePlanId(null); }}
          onCancelOpen={() => setShowCancel(true)}
          onCancelClose={() => setShowCancel(false)}
          onCancelConfirm={handleCancel}
        />
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="text-center">
        <h2 className="text-2xl font-bold text-foreground">{t('plans.title')}</h2>
        <p className="mt-2 text-muted-foreground">{t('noSubscriptionDesc')}</p>
      </div>
      <PlanCards
        plans={plans}
        currentTier={null}
        onSelectPlan={handleSelectPlan}
        isLoading={isLoadingPlans}
      />
      {showPayment && selectedPlan && (
        <PaymentForm
          planName={selectedPlan.name}
          amount={selectedPlan.priceMonthly}
          onSubmit={handleSubscribe}
          onCancel={() => { setShowPayment(false); setSelectedPlanId(null); }}
          isSubmitting={isSubscribing}
        />
      )}
    </div>
  );
}
