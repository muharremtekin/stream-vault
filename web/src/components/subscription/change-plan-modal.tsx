'use client';

import { useState } from 'react';

import { ArrowRight, CreditCard } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { formatCurrency } from '@/lib/utils/format';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Modal } from '@/components/ui/modal';
import type { Plan } from '@/lib/types/subscription';

interface ChangePlanModalProps {
  isOpen: boolean;
  onClose: () => void;
  currentPlan: Plan;
  newPlan: Plan;
  onConfirm: (cardNumber: string) => void;
  isChanging: boolean;
}

export function ChangePlanModal({
  isOpen,
  onClose,
  currentPlan,
  newPlan,
  onConfirm,
  isChanging,
}: ChangePlanModalProps) {
  const t = useTranslations('subscription.changePlan');
  const tCommon = useTranslations('common');
  const tPayment = useTranslations('subscription.payment');
  const [cardNumber, setCardNumber] = useState('');
  const [cardError, setCardError] = useState('');

  const priceDiff = newPlan.priceMonthly - currentPlan.priceMonthly;
  const isUpgrade = priceDiff > 0;

  const handleConfirm = () => {
    if (!/^\d{16}$/.test(cardNumber)) {
      setCardError(tPayment('cardNumberInvalid'));
      return;
    }
    setCardError('');
    onConfirm(cardNumber);
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={t('title')} size="md">
      <div className="space-y-4">
        <div className="flex items-center justify-center gap-4 rounded-md bg-muted/50 p-4">
          <div className="text-center">
            <p className="text-xs text-muted-foreground">{t('from')}</p>
            <p className="text-lg font-semibold text-foreground">{currentPlan.name}</p>
            <p className="text-sm text-muted-foreground">
              {formatCurrency(currentPlan.priceMonthly)}
            </p>
          </div>
          <ArrowRight className="h-5 w-5 text-muted-foreground" />
          <div className="text-center">
            <p className="text-xs text-muted-foreground">{t('to')}</p>
            <p className="text-lg font-semibold text-primary">{newPlan.name}</p>
            <p className="text-sm text-muted-foreground">
              {formatCurrency(newPlan.priceMonthly)}
            </p>
          </div>
        </div>

        <div className="rounded-md border border-border p-3">
          <p className="text-sm font-medium text-foreground">
            {t('priceDifference')}: {formatCurrency(Math.abs(priceDiff))}
          </p>
          <p className="mt-1 text-xs text-muted-foreground">
            {isUpgrade
              ? t('upgradeNote', { amount: formatCurrency(Math.abs(priceDiff)) })
              : t('downgradeNote', { amount: formatCurrency(Math.abs(priceDiff)) })}
          </p>
        </div>

        <div>
          <p className="mb-2 text-sm text-muted-foreground">{t('enterCard')}</p>
          <Input
            placeholder="1234567890123456"
            leftIcon={<CreditCard className="h-4 w-4" />}
            maxLength={16}
            value={cardNumber}
            onChange={(e) => {
              setCardNumber(e.target.value);
              setCardError('');
            }}
            error={cardError}
          />
        </div>

        <div className="flex gap-3 pt-2">
          <Button
            onClick={handleConfirm}
            isLoading={isChanging}
            className="flex-1"
          >
            {t('confirm')}
          </Button>
          <Button variant="ghost" onClick={onClose}>
            {tCommon('cancel')}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
