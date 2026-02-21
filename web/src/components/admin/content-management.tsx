'use client';

import { useState } from 'react';

import { ContentToolbar } from '@/components/admin/content-toolbar';
import { ContentTable } from '@/components/admin/content-table';

export function ContentManagement() {
  const [search, setSearch] = useState('');

  return (
    <div className="space-y-4">
      <ContentToolbar search={search} onSearchChange={setSearch} />
      <ContentTable search={search} />
    </div>
  );
}
