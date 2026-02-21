'use client';

import { useState } from 'react';

import { Menu } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { Sidebar } from '@/components/layout/sidebar';
import { AdminMobileSidebar } from '@/components/layout/admin-mobile-sidebar';

interface AdminLayoutShellProps {
  children: React.ReactNode;
}

export function AdminLayoutShell({ children }: AdminLayoutShellProps) {
  const t = useTranslations('layout.admin');
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);

  return (
    <div className="flex">
      <Sidebar />
      <div className="min-h-[calc(100vh-4rem)] w-full lg:pl-56">
        <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
          <div className="mb-4 lg:hidden">
            <button
              onClick={() => setIsSidebarOpen(true)}
              aria-label={t('openSidebar')}
              className="rounded-md border border-border bg-card p-2 text-muted-foreground transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
            >
              <Menu className="h-5 w-5" />
            </button>
          </div>
          {children}
        </div>
      </div>
      <AdminMobileSidebar
        isOpen={isSidebarOpen}
        onClose={() => setIsSidebarOpen(false)}
      />
    </div>
  );
}
