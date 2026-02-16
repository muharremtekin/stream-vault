export interface Plan {
  id: string;
  name: string;
  tier: string;
  priceMonthly: number;
  maxScreens: number;
  maxQuality: string;
  features: string;
  isActive: boolean;
}

export interface Subscription {
  id: string;
  userId: string;
  plan: Plan;
  status: string;
  periodStart: string;
  periodEnd: string;
  autoRenew: boolean;
  cancelledAt: string | null;
  createdAt: string;
}

export interface Invoice {
  id: string;
  subscriptionId: string;
  invoiceNumber: string;
  amount: number;
  currency: string;
  periodStart: string;
  periodEnd: string;
  issuedAt: string;
}

export interface Payment {
  id: string;
  subscriptionId: string;
  amount: number;
  currency: string;
  status: string;
  transactionId: string | null;
  cardLastFour: string;
  createdAt: string;
}

export interface CreateSubscriptionRequest {
  planId: string;
  cardNumber: string;
}

export interface CreateSubscriptionResult {
  subscriptionId: string;
  userId: string;
  planId: string;
  planName: string;
  tier: string;
  status: string;
  periodStart: string;
  periodEnd: string;
  amountCharged: number;
  currency: string;
}

export interface ChangePlanRequest {
  newPlanId: string;
  cardNumber: string;
}

export interface ChangePlanResult {
  subscriptionId: string;
  oldPlanId: string;
  oldPlanName: string;
  newPlanId: string;
  newPlanName: string;
  newTier: string;
  priceDifference: number;
  status: string;
}

export interface CancelSubscriptionResult {
  subscriptionId: string;
  status: string;
  cancelledAt: string;
  periodEnd: string;
}
