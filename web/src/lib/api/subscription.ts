import { apiClient } from '@/lib/api/client';
import type {
  Plan,
  Subscription,
  Invoice,
  Payment,
  CreateSubscriptionRequest,
  CreateSubscriptionResult,
  ChangePlanRequest,
  ChangePlanResult,
  CancelSubscriptionResult,
} from '@/lib/types/subscription';

export async function getPlans(): Promise<Plan[]> {
  const response = await apiClient.get<Plan[]>('/api/plans');
  return response.data;
}

export async function subscribe(
  data: CreateSubscriptionRequest
): Promise<CreateSubscriptionResult> {
  const response = await apiClient.post<CreateSubscriptionResult>(
    '/api/subscriptions',
    data
  );
  return response.data;
}

export async function getMySubscription(): Promise<Subscription> {
  const response = await apiClient.get<Subscription>('/api/subscriptions/me');
  return response.data;
}

export async function changePlan(
  data: ChangePlanRequest
): Promise<ChangePlanResult> {
  const response = await apiClient.put<ChangePlanResult>(
    '/api/subscriptions/me/plan',
    data
  );
  return response.data;
}

export async function cancelSubscription(): Promise<CancelSubscriptionResult> {
  const response = await apiClient.post<CancelSubscriptionResult>(
    '/api/subscriptions/me/cancel'
  );
  return response.data;
}

export async function getInvoices(): Promise<Invoice[]> {
  const response = await apiClient.get<Invoice[]>(
    '/api/subscriptions/me/invoices'
  );
  return response.data;
}

export async function getPayments(
  limit?: number,
  offset?: number
): Promise<Payment[]> {
  const response = await apiClient.get<Payment[]>(
    '/api/subscriptions/me/payments',
    { params: { ...(limit ? { limit } : {}), ...(offset ? { offset } : {}) } }
  );
  return response.data;
}
