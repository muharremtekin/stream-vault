'use client';

import { useEffect } from 'react';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { X } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';

interface MobileMenuProps {
  isOpen: boolean;
  onClose: () => void;
}

const NAV_ITEMS = [
  { href: '/browse', key: 'home' },
  { href: '/series', key: 'series' },
  { href: '/movies', key: 'movies' },
  { href: '/my-list', key: 'myList' },
] as const;

export function MobileMenu({ isOpen, onClose }: MobileMenuProps) {
  const t = useTranslations('layout.navbar');
  const pathname = usePathname();

  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }

    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 md:hidden">
      <div
        className="absolute inset-0 bg-background/80 backdrop-blur-sm"
        onClick={onClose}
        onKeyDown={(e) => {
          if (e.key === 'Escape') onClose();
        }}
        role="button"
        tabIndex={-1}
        aria-label={t('closeMenu')}
      />
      <nav className="absolute left-0 top-0 h-full w-64 border-r border-border bg-card p-6 shadow-lg">
        <div className="mb-8 flex items-center justify-between">
          <span className="text-xl font-bold text-primary">StreamVault</span>
          <button
            onClick={onClose}
            aria-label={t('closeMenu')}
            className="rounded-full p-1 text-muted-foreground transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
          >
            <X className="h-5 w-5" />
          </button>
        </div>
        <ul className="flex flex-col gap-2">
          {NAV_ITEMS.map((item) => (
            <li key={item.href}>
              <Link
                href={item.href}
                onClick={onClose}
                className={cn(
                  'block rounded-md px-3 py-2 text-sm font-medium transition-colors',
                  pathname === item.href
                    ? 'bg-muted text-foreground'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                )}
              >
                {t(item.key)}
              </Link>
            </li>
          ))}
        </ul>
      </nav>
    </div>
  );
}
