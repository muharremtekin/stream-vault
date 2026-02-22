'use client';

import { useState, useRef, useEffect } from 'react';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { User, ChevronDown, Shield } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';
import { useAuthStore } from '@/lib/stores/auth-store';
import { SUBSCRIPTION_TIERS } from '@/lib/types/common';

export function ProfileSwitcher() {
  const t = useTranslations('layout');
  const router = useRouter();
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const isAdmin = user?.role === SUBSCRIPTION_TIERS.Admin;
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen]);

  return (
    <div ref={dropdownRef} className="relative">
      <button
        onClick={() => setIsOpen((prev) => !prev)}
        aria-label={t('navbar.profile')}
        className="flex items-center gap-1 rounded-full p-1 text-muted-foreground transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
      >
        <div className="flex h-8 w-8 items-center justify-center rounded bg-muted">
          <User className="h-5 w-5" />
        </div>
        <ChevronDown
          className={cn(
            'h-4 w-4 transition-transform',
            isOpen && 'rotate-180'
          )}
        />
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-48 rounded-md border border-border bg-card py-1 shadow-lg">
          <Link
            href="/account/profiles"
            onClick={() => setIsOpen(false)}
            className="block px-4 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          >
            {t('profileSwitcher.switchProfile')}
          </Link>
          <Link
            href="/account/profiles"
            onClick={() => setIsOpen(false)}
            className="block px-4 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          >
            {t('profileSwitcher.manageProfiles')}
          </Link>
          <Link
            href="/account"
            onClick={() => setIsOpen(false)}
            className="block px-4 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          >
            {t('navbar.account')}
          </Link>
          {isAdmin && (
            <Link
              href="/admin"
              onClick={() => setIsOpen(false)}
              className="flex items-center gap-2 px-4 py-2 text-sm text-primary transition-colors hover:bg-muted hover:text-foreground"
            >
              <Shield className="h-4 w-4" />
              {t('admin.admin')}
            </Link>
          )}
          <div className="my-1 border-t border-border" />
          <button
            onClick={() => {
              setIsOpen(false);
              logout();
              router.push('/login');
            }}
            className="block w-full px-4 py-2 text-left text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          >
            {t('navbar.signOut')}
          </button>
        </div>
      )}
    </div>
  );
}
