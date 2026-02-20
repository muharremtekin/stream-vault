'use client';

import { AlertTriangle } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { formatDate } from '@/lib/utils/format';
import { Button } from '@/components/ui/button';
import { Modal } from '@/components/ui/modal';

interface CancelModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  isCancelling: boolean;
  periodEnd: string;
}

export function CancelModal({
  isOpen,
  onClose,
  onConfirm,
  isCancelling,
  periodEnd,
}: CancelModalProps) {
  const t = useTranslations('subscription.cancel');

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={t('title')} size="sm">
      <div className="space-y-4">
        <div className="flex items-start gap-3 rounded-md bg-error/10 p-4">
          <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0 text-error" />
          <div>
            <p className="font-medium text-foreground">{t('warning')}</p>
            <p className="mt-1 text-sm text-muted-foreground">
              {t('description', { date: formatDate(periodEnd) })}
            </p>
          </div>
        </div>

        <div className="flex gap-3 pt-2">
          <Button
            variant="danger"
            onClick={onConfirm}
            isLoading={isCancelling}
            className="flex-1"
          >
            {t('confirm')}
          </Button>
          <Button variant="secondary" onClick={onClose}>
            {t('keepPlan')}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
