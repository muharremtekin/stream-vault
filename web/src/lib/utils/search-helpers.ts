import type { SearchHit, TrendingItem, SearchParams } from '@/lib/types/search';
import type { ContentCardItem } from '@/lib/types/browse';

export function searchHitToCardItem(hit: SearchHit): ContentCardItem {
  return {
    contentId: hit.contentId,
    title: hit.title,
    thumbnailUrl: hit.thumbnailUrl || undefined,
    releaseYear: hit.releaseYear,
    averageRating: hit.averageRating,
    genres: hit.genres,
    contentType: hit.contentType,
  };
}

export function trendingItemToCardItem(item: TrendingItem): ContentCardItem {
  return {
    contentId: item.contentId,
    title: item.title,
    thumbnailUrl: item.thumbnailUrl || undefined,
    releaseYear: item.releaseYear,
    averageRating: item.averageRating,
    contentType: item.contentType,
    rank: item.rank,
  };
}

export function buildSearchParams(urlParams: URLSearchParams): SearchParams {
  const q = urlParams.get('q') ?? undefined;
  const genres = urlParams.get('genres') ?? undefined;
  const content_type = urlParams.get('content_type') ?? undefined;
  const yearFrom = urlParams.get('year_from');
  const yearTo = urlParams.get('year_to');
  const minRating = urlParams.get('min_rating');
  const sort = urlParams.get('sort') ?? undefined;
  const order = urlParams.get('order') ?? undefined;
  const pageStr = urlParams.get('page');

  return {
    q,
    genres,
    content_type,
    year_from: yearFrom ? Number(yearFrom) : undefined,
    year_to: yearTo ? Number(yearTo) : undefined,
    min_rating: minRating ? Number(minRating) : undefined,
    sort,
    order,
    page: pageStr ? Number(pageStr) : 1,
    pageSize: 24,
  };
}
