import { Suspense } from 'react';

import { SearchResults } from '@/components/search/search-results';

import SearchLoading from './loading';

export default function SearchPage() {
  return (
    <div className="px-4 py-8 sm:px-8 lg:px-12">
      <Suspense fallback={<SearchLoading />}>
        <SearchResults />
      </Suspense>
    </div>
  );
}
