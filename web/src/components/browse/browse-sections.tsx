'use client';

import { useTranslations } from 'next-intl';

import ContentCard from '@/components/browse/content-card';
import ContentRow from '@/components/browse/content-row';
import ContinueWatchingRow from '@/components/browse/continue-watching-row';
import { useHomeSections, useTrending } from '@/lib/hooks/use-browse';
import type { ContentCardItem } from '@/lib/types/browse';
import type { HomePageSection } from '@/lib/types/recommendation';
import type { TrendingItem } from '@/lib/types/search';

const SECTION_TYPE_KEY: Record<string, string> = {
  personal: 'forYou',
  trending: 'trending',
  because_you_watched: 'becauseYouWatched',
  new: 'newReleases',
};

function sectionTitle(section: HomePageSection, t: (key: string) => string): string {
  const key = SECTION_TYPE_KEY[section.sectionType];
  return key ? t(key) : section.title;
}

function trendingToCard(item: TrendingItem): ContentCardItem {
  return {
    contentId: item.contentId,
    title: item.title,
    thumbnailUrl: item.thumbnailUrl,
    releaseYear: item.releaseYear,
    averageRating: item.averageRating,
    contentType: item.contentType,
    rank: item.rank,
  };
}

export default function BrowseSections() {
  const t = useTranslations('browse');
  const { data: homeData, isLoading: loadingHome } = useHomeSections();
  const { data: trendingData, isLoading: loadingTrending } = useTrending();

  const sections = homeData?.sections ?? [];
  const trendingItems = trendingData?.items ?? [];

  return (
    <div className="space-y-8 py-4">
      {/* Continue Watching — always first */}
      <ContinueWatchingRow />

      {/* Recommendation sections */}
      {loadingHome
        ? Array.from({ length: 3 }).map((_, i) => (
            // eslint-disable-next-line react/no-array-index-key
            <ContentRow key={`rec-skeleton-${i}`} title="..." isLoading />
          ))
        : sections.map((section) => {
            if (section.items.length === 0) return null;
            return (
              <ContentRow key={section.sectionType} title={sectionTitle(section, t)}>
                {section.items.map((item) => (
                  <ContentCard
                    key={item.contentId}
                    item={{
                      contentId: item.contentId,
                      title: item.title,
                      thumbnailUrl: item.thumbnailUrl,
                      releaseYear: item.releaseYear,
                      averageRating: item.averageRating,
                      contentType: item.contentType,
                    }}
                  />
                ))}
              </ContentRow>
            );
          })}

      {/* Trending row (numbered) */}
      <ContentRow title={t('trending')} isLoading={loadingTrending}>
        {trendingItems.map((item) => (
          <ContentCard
            key={item.contentId}
            item={trendingToCard(item)}
            showRank
          />
        ))}
      </ContentRow>
    </div>
  );
}
