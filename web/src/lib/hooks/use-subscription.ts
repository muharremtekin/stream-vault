'use client';

import { useEffect } from 'react';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

import {
  getPlans,
  getMySubscription,
  subscribe,
  changePlan,
  cancelSubscription,
  getInvoices,
} from '@/lib/api/subscription';
import { useAuthStore } from '@/lib/stores/auth-store';
import type { SubscriptionTier } from '@/lib/types/common';
import type {
  CreateSubscriptionRequest,
  ChangePlanRequest,
} from '@/lib/types/subscription';

export function useSubscription() {
  const queryClient = useQueryClient();
  const isAuthenticated = useAuthStore((s) => s.accessToken !== null);
  const setSubscriptionTier = useAuthStore((s) => s.setSubscriptionTier);

  const plansQuery = useQuery({
    queryKey: ['subscription', 'plans'],
    queryFn: getPlans,
    staleTime: 10 * 60_000,
  });

  const subscriptionQuery = useQuery({
    queryKey: ['subscription', 'me'],
    queryFn: getMySubscription,
    enabled: isAuthenticated,
    staleTime: 2 * 60_000,
  });

  // Sync subscription tier to auth store
  useEffect(() => {
    if (subscriptionQuery.data?.plan?.tier) {
      setSubscriptionTier(
        subscriptionQuery.data.plan.tier as SubscriptionTier
      );
    }
  }, [subscriptionQuery.data, setSubscriptionTier]);

  const subscribeMutation = useMutation({
    mutationFn: (data: CreateSubscriptionRequest) => subscribe(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['subscription'] });
    },
  });

  const changePlanMutation = useMutation({
    mutationFn: (data: ChangePlanRequest) => changePlan(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['subscription'] });
    },
  });

  const cancelMutation = useMutation({
    mutationFn: cancelSubscription,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['subscription'] });
    },
  });

  const invoicesQuery = useQuery({
    queryKey: ['subscription', 'invoices'],
    queryFn: getInvoices,
    enabled: isAuthenticated,
    staleTime: 5 * 60_000,
  });

  return {
    plans: plansQuery.data ?? [],
    subscription: subscriptionQuery.data,
    invoices: invoicesQuery.data ?? [],
    isLoadingPlans: plansQuery.isLoading,
    isLoadingSubscription: subscriptionQuery.isLoading,
    subscribe: subscribeMutation.mutateAsync,
    changePlan: changePlanMutation.mutateAsync,
    cancel: cancelMutation.mutateAsync,
    isSubscribing: subscribeMutation.isPending,
    isChangingPlan: changePlanMutation.isPending,
    isCancelling: cancelMutation.isPending,
  };
}
