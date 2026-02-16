import Link from 'next/link';
import { getTranslations } from 'next-intl/server';

export async function Footer() {
  const t = await getTranslations('layout.footer');
  const year = new Date().getFullYear();

  return (
    <footer className="border-t border-border bg-background">
      <div className="mx-auto max-w-7xl px-4 py-8">
        <div className="flex flex-col items-center justify-between gap-4 md:flex-row">
          <nav className="flex flex-wrap justify-center gap-6 text-sm">
            <Link
              href="/about"
              className="text-muted-foreground transition-colors hover:text-foreground"
            >
              {t('about')}
            </Link>
            <Link
              href="/help"
              className="text-muted-foreground transition-colors hover:text-foreground"
            >
              {t('help')}
            </Link>
            <Link
              href="/terms"
              className="text-muted-foreground transition-colors hover:text-foreground"
            >
              {t('terms')}
            </Link>
            <Link
              href="/privacy"
              className="text-muted-foreground transition-colors hover:text-foreground"
            >
              {t('privacy')}
            </Link>
          </nav>
          <p className="text-sm text-muted-foreground">
            {t('copyright', { year })}
          </p>
        </div>
      </div>
    </footer>
  );
}
