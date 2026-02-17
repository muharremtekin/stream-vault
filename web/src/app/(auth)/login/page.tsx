import Link from 'next/link';
import { getTranslations } from 'next-intl/server';

import { LoginForm } from '@/components/auth/login-form';

export default async function LoginPage() {
  const t = await getTranslations('auth');

  return (
    <div className="w-full max-w-md rounded-lg bg-card p-8">
      <h1 className="mb-6 text-2xl font-bold text-foreground">
        {t('loginTitle')}
      </h1>
      <LoginForm />
      <p className="mt-6 text-center text-sm text-muted-foreground">
        {t('noAccount')}{' '}
        <Link
          href="/register"
          className="font-medium text-primary hover:underline"
        >
          {t('register')}
        </Link>
      </p>
    </div>
  );
}
