import type { Movie } from '@/lib/types/catalog';
import type { ContentCardItem } from '@/lib/types/browse';
import type { CatalogParams } from '@/lib/types/catalog';

export function movieToCardItem(movie: Movie): ContentCardItem {
  return {
    contentId: movie.id,
    title: movie.title,
    thumbnailUrl: movie.thumbnailUrl || undefined,
    releaseYear: movie.releaseYear,
    averageRating: movie.averageRating,
    genres: movie.genres,
    contentType: 'Movie',
    durationMinutes: movie.durationMinutes,
  };
}

export function buildMovieParams(urlParams: URLSearchParams): CatalogParams {
  const genre = urlParams.get('genre') ?? undefined;
  const sort = urlParams.get('sort') ?? undefined;
  const pageStr = urlParams.get('page');

  return {
    genre,
    sort,
    page: pageStr ? Number(pageStr) : 1,
    pageSize: 24,
  };
}
