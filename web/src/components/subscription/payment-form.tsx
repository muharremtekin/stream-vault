'use client';

import { useForm } from 'react-hook-form';

import { zodResolver } from '@hookform/resolvers/zod';
import { CreditCard, Calendar, Lock } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { formatCurrency } from '@/lib/utils/format';
import { paymentFormSchema } from '@/lib/validations/subscription';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import type { PaymentFormData } from '@/lib/validations/subscription';

interface PaymentFormProps {
  planName: string;
  amount: number;
  onSubmit: (cardNumber: string) => void;
  onCancel: () => void;
  isSubmitting: boolean;
}

export function PaymentForm({
  planName,
  amount,
  onSubmit,
  onCancel,
  isSubmitting,
}: PaymentFormProps) {
  const t = useTranslations('subscription.payment');
  const tCommon = useTranslations('common');

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<PaymentFormData>({
    resolver: zodResolver(paymentFormSchema),
    defaultValues: { cardNumber: '', expiry: '', cvv: '' },
  });

  const handleFormSubmit = (data: PaymentFormData) => {
    onSubmit(data.cardNumber);
  };

  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <h3 className="mb-4 text-lg font-semibold text-foreground">{t('title')}</h3>

      <div className="mb-6 rounded-md bg-muted/50 p-4">
        <p className="text-sm text-muted-foreground">{t('summary')}</p>
        <div className="mt-1 flex items-center justify-between">
          <span className="font-medium text-foreground">{planName}</span>
          <span className="text-lg font-bold text-foreground">{formatCurrency(amount)}</span>
        </div>
      </div>

      <form onSubmit={handleSubmit(handleFormSubmit)} noValidate className="space-y-4">
        <Input
          label={t('cardNumber')}
          placeholder={t('cardNumberPlaceholder')}
          leftIcon={<CreditCard className="h-4 w-4" />}
          error={errors.cardNumber?.message}
          maxLength={16}
          {...register('cardNumber')}
        />
        <div className="grid grid-cols-2 gap-4">
          <Input
            label={t('expiry')}
            placeholder={t('expiryPlaceholder')}
            leftIcon={<Calendar className="h-4 w-4" />}
            error={errors.expiry?.message}
            maxLength={5}
            {...register('expiry')}
          />
          <Input
            label={t('cvv')}
            placeholder={t('cvvPlaceholder')}
            leftIcon={<Lock className="h-4 w-4" />}
            error={errors.cvv?.message}
            maxLength={4}
            type="password"
            {...register('cvv')}
          />
        </div>
        <div className="flex gap-3 pt-2">
          <Button type="submit" isLoading={isSubmitting} className="flex-1">
            {t('confirmPayment')}
          </Button>
          <Button type="button" variant="ghost" onClick={onCancel}>
            {tCommon('cancel')}
          </Button>
        </div>
      </form>
    </div>
  );
}
