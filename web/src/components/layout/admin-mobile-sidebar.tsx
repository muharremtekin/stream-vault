'use client';

import { useEffect } from 'react';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { X, LayoutDashboard, Film, Settings } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';

const SIDEBAR_ITEMS = [
  { href: '/admin', key: 'dashboard', icon: LayoutDashboard },
  { href: '/admin/content', key: 'content', icon: Film },
  { href: '/admin/encoding', key: 'encoding', icon: Settings },
] as const;

interface AdminMobileSidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

export function AdminMobileSidebar({ isOpen, onClose }: AdminMobileSidebarProps) {
  const t = useTranslations('layout.admin');
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
    <div className="fixed inset-0 z-50 lg:hidden">
      <div
        className="absolute inset-0 bg-background/80 backdrop-blur-sm animate-overlay-fade-in"
        onClick={onClose}
        onKeyDown={(e) => {
          if (e.key === 'Escape') onClose();
        }}
        role="button"
        tabIndex={-1}
        aria-label={t('closeSidebar')}
      />
      <nav className="absolute left-0 top-0 h-full w-56 border-r border-border bg-card p-4 shadow-lg animate-slide-in-from-left">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="px-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            {t('admin')}
          </h2>
          <button
            onClick={onClose}
            aria-label={t('closeSidebar')}
            className="rounded-full p-1 text-muted-foreground transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
          >
            <X className="h-5 w-5" />
          </button>
        </div>
        <div className="flex flex-col gap-1">
          {SIDEBAR_ITEMS.map((item) => {
            const isActive =
              item.href === '/admin'
                ? pathname === '/admin'
                : pathname.startsWith(item.href);

            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={onClose}
                className={cn(
                  'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-muted text-foreground'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                )}
              >
                <item.icon className="h-4 w-4" />
                {t(item.key)}
              </Link>
            );
          })}
        </div>
      </nav>
    </div>
  );
}
