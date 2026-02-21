'use client';

import Link from 'next/link';
import { Search, Plus } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

interface ContentToolbarProps {
  search: string;
  onSearchChange: (value: string) => void;
}

export function ContentToolbar({ search, onSearchChange }: ContentToolbarProps) {
  const t = useTranslations('admin.content');

  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <Input
        placeholder={t('searchPlaceholder')}
        value={search}
        onChange={(e) => onSearchChange(e.target.value)}
        leftIcon={<Search className="h-4 w-4" />}
        className="sm:max-w-xs"
      />
      <Link href="/admin/content/new">
        <Button size="sm">
          <Plus className="h-4 w-4" />
          {t('addNew')}
        </Button>
      </Link>
    </div>
  );
}
