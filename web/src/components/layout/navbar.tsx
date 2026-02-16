'use client';

import { useState, useEffect } from 'react';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Menu } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { cn } from '@/lib/utils/cn';
import { SearchButton } from '@/components/layout/search-button';
import { NotificationBell } from '@/components/layout/notification-bell';
import { ProfileSwitcher } from '@/components/layout/profile-switcher';
import { MobileMenu } from '@/components/layout/mobile-menu';

const NAV_ITEMS = [
  { href: '/browse', key: 'home' },
  { href: '/series', key: 'series' },
  { href: '/movies', key: 'movies' },
  { href: '/my-list', key: 'myList' },
] as const;

export function Navbar() {
  const t = useTranslations('layout.navbar');
  const pathname = usePathname();
  const [isScrolled, setIsScrolled] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  useEffect(() => {
    function handleScroll() {
      setIsScrolled(window.scrollY > 10);
    }

    window.addEventListener('scroll', handleScroll, { passive: true });
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  return (
    <>
      <nav
        className={cn(
          'fixed top-0 z-40 w-full transition-colors duration-300',
          isScrolled
            ? 'bg-background/95 shadow-md backdrop-blur-sm'
            : 'bg-gradient-to-b from-background/80 to-transparent'
        )}
      >
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4">
          <div className="flex items-center gap-8">
            <button
              onClick={() => setIsMobileMenuOpen(true)}
              aria-label={t('openMenu')}
              className="rounded-full p-2 text-muted-foreground transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none md:hidden"
            >
              <Menu className="h-5 w-5" />
            </button>

            <Link
              href="/browse"
              className="text-xl font-bold tracking-tight text-primary"
            >
              StreamVault
            </Link>

            <ul className="hidden items-center gap-6 md:flex">
              {NAV_ITEMS.map((item) => (
                <li key={item.href}>
                  <Link
                    href={item.href}
                    className={cn(
                      'text-sm font-medium transition-colors',
                      pathname === item.href
                        ? 'text-foreground'
                        : 'text-muted-foreground hover:text-foreground'
                    )}
                  >
                    {t(item.key)}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          <div className="flex items-center gap-2">
            <SearchButton />
            <NotificationBell />
            <ProfileSwitcher />
          </div>
        </div>
      </nav>

      <MobileMenu
        isOpen={isMobileMenuOpen}
        onClose={() => setIsMobileMenuOpen(false)}
      />
    </>
  );
}
