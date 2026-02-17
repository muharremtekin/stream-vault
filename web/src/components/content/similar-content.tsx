'use client';

import { useTranslations } from 'next-intl';

import ContentCard from '@/components/browse/content-card';
import ContentRow from '@/components/browse/content-row';
import { useSimilarContent } from '@/lib/hooks/use-content';
import type { ContentCardItem } from '@/lib/types/browse';
import type { SimilarItem } from '@/lib/types/recommendation';

interface SimilarContentProps {
  contentId: string;
}

function toCardItem(item: SimilarItem): ContentCardItem {
  return {
    contentId: item.contentId,
    title: item.title,
    thumbnailUrl: item.thumbnailUrl,
    releaseYear: item.releaseYear,
    averageRating: item.averageRating,
    contentType: item.contentType,
  };
}

export default function SimilarContent({ contentId }: SimilarContentProps) {
  const t = useTranslations('browse');
  const { data, isLoading } = useSimilarContent(contentId);

  const items = data?.items ?? [];

  if (!isLoading && items.length === 0) return null;

  return (
    <ContentRow title={t('similar')} isLoading={isLoading}>
      {items.map((item) => (
        <ContentCard key={item.contentId} item={toCardItem(item)} />
      ))}
    </ContentRow>
  );
}
