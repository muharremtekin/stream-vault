/**
 * Format duration in minutes to Turkish "Xs Ydk" format.
 * e.g. 120 → "2s 0dk", 95 → "1s 35dk", 45 → "45dk"
 */
export function formatDuration(minutes: number): string {
  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;
  if (hours === 0) return `${mins}dk`;
  return `${hours}s ${mins}dk`;
}

/**
 * Format date to Turkish locale string.
 * e.g. "2024-01-15" → "15 Ocak 2024"
 */
export function formatDate(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleDateString('tr-TR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

/**
 * Format date as relative time (e.g. "3 gün önce", "az önce").
 */
export function formatRelativeTime(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffSec = Math.floor(diffMs / 1000);
  const diffMin = Math.floor(diffSec / 60);
  const diffHour = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHour / 24);

  if (diffSec < 60) return 'az önce';
  if (diffMin < 60) return `${diffMin}dk önce`;
  if (diffHour < 24) return `${diffHour}s önce`;
  if (diffDay < 7) return `${diffDay} gün önce`;
  return formatDate(d);
}

/**
 * Format currency in Turkish Lira.
 * e.g. 79.99 → "₺79,99"
 */
export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('tr-TR', {
    style: 'currency',
    currency: 'TRY',
    minimumFractionDigits: 2,
  }).format(amount);
}
