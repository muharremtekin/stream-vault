import Link from 'next/link';
import { getTranslations } from 'next-intl/server';

import { RegisterForm } from '@/components/auth/register-form';

export default async function RegisterPage() {
  const t = await getTranslations('auth');

  return (
    <div className="w-full max-w-md rounded-lg bg-card p-8">
      <h1 className="mb-6 text-2xl font-bold text-foreground">
        {t('registerTitle')}
      </h1>
      <RegisterForm />
      <p className="mt-6 text-center text-sm text-muted-foreground">
        {t('hasAccount')}{' '}
        <Link
          href="/login"
          className="font-medium text-primary hover:underline"
        >
          {t('login')}
        </Link>
      </p>
    </div>
  );
}
