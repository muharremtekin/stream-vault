'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { LayoutDashboard, Film, Settings } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';

import type { ReactNode } from 'react';

const SIDEBAR_ITEMS = [
  { href: '/admin', key: 'dashboard', icon: LayoutDashboard },
  { href: '/admin/content', key: 'content', icon: Film },
  { href: '/admin/encoding', key: 'encoding', icon: Settings },
] as const;

interface SidebarLinkProps {
  href: string;
  isActive: boolean;
  icon: ReactNode;
  children: ReactNode;
}

function SidebarLink({ href, isActive, icon, children }: SidebarLinkProps) {
  return (
    <Link
      href={href}
      className={cn(
        'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
        isActive
          ? 'bg-muted text-foreground'
          : 'text-muted-foreground hover:bg-muted hover:text-foreground'
      )}
    >
      {icon}
      {children}
    </Link>
  );
}

export function Sidebar() {
  const t = useTranslations('layout.admin');
  const pathname = usePathname();

  return (
    <aside className="fixed left-0 top-16 hidden h-[calc(100vh-4rem)] w-56 border-r border-border bg-card lg:block">
      <div className="p-4">
        <h2 className="mb-4 px-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          {t('admin')}
        </h2>
        <nav className="flex flex-col gap-1">
          {SIDEBAR_ITEMS.map((item) => {
            const isActive =
              item.href === '/admin'
                ? pathname === '/admin'
                : pathname.startsWith(item.href);

            return (
              <SidebarLink
                key={item.href}
                href={item.href}
                isActive={isActive}
                icon={<item.icon className="h-4 w-4" />}
              >
                {t(item.key)}
              </SidebarLink>
            );
          })}
        </nav>
      </div>
    </aside>
  );
}
