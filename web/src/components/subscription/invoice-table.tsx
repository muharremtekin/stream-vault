'use client';

import { FileText } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { formatCurrency, formatDate } from '@/lib/utils/format';
import { Skeleton } from '@/components/ui/skeleton';
import type { Invoice } from '@/lib/types/subscription';

interface InvoiceTableProps {
  invoices: Invoice[];
  isLoading: boolean;
}

export function InvoiceTable({ invoices, isLoading }: InvoiceTableProps) {
  const t = useTranslations('subscription.invoices');

  if (isLoading) {
    return (
      <div className="rounded-lg border border-border bg-card p-6">
        <Skeleton variant="text" className="mb-4 h-6 w-40" />
        <div className="space-y-3">
          {['skeleton-inv-1', 'skeleton-inv-2', 'skeleton-inv-3'].map((id) => (
            <Skeleton key={id} variant="text" className="h-10 w-full" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <h3 className="mb-4 text-lg font-semibold text-foreground">{t('title')}</h3>

      {invoices.length === 0 ? (
        <div className="flex flex-col items-center py-8 text-center">
          <FileText className="mb-3 h-10 w-10 text-muted-foreground/50" />
          <p className="text-sm font-medium text-foreground">{t('noInvoices')}</p>
          <p className="mt-1 text-xs text-muted-foreground">{t('noInvoicesDesc')}</p>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-left">
                <th scope="col" className="pb-3 pr-4 font-medium text-muted-foreground">
                  {t('date')}
                </th>
                <th scope="col" className="pb-3 pr-4 font-medium text-muted-foreground">
                  {t('invoiceNumber')}
                </th>
                <th scope="col" className="pb-3 pr-4 font-medium text-muted-foreground">
                  {t('period')}
                </th>
                <th scope="col" className="pb-3 text-right font-medium text-muted-foreground">
                  {t('amount')}
                </th>
              </tr>
            </thead>
            <tbody>
              {invoices.map((invoice) => (
                <tr key={invoice.id} className="border-b border-border/50 last:border-0">
                  <td className="py-3 pr-4 text-foreground">
                    {formatDate(invoice.issuedAt)}
                  </td>
                  <td className="py-3 pr-4 font-mono text-xs text-muted-foreground">
                    {invoice.invoiceNumber}
                  </td>
                  <td className="py-3 pr-4 text-muted-foreground">
                    {formatDate(invoice.periodStart)} – {formatDate(invoice.periodEnd)}
                  </td>
                  <td className="py-3 text-right font-medium text-foreground">
                    {formatCurrency(invoice.amount)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
