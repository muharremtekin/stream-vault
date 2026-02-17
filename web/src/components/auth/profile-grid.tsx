'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

import { Plus } from 'lucide-react';
import { useTranslations } from 'next-intl';

import { useProfile } from '@/lib/hooks/use-profile';
import { cn } from '@/lib/utils/cn';
import { MAX_PROFILES } from '@/lib/utils/constants';
import { Skeleton } from '@/components/ui/skeleton';
import { ProfileCard } from '@/components/auth/profile-card';
import { CreateProfileModal } from '@/components/auth/create-profile-modal';
import type { Profile } from '@/lib/types/auth';

export function ProfileGrid() {
  const t = useTranslations('auth');
  const router = useRouter();
  const { profiles, isLoading, switchProfile } = useProfile();
  const [isModalOpen, setIsModalOpen] = useState(false);

  const handleSelectProfile = (profile: Profile) => {
    switchProfile(profile);
    router.push('/browse');
  };

  if (isLoading) {
    return (
      <div className="flex flex-col items-center">
        <Skeleton className="mb-8 h-10 w-64" />
        <div className="flex gap-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={`skel-${i}`} className="flex flex-col items-center gap-2">
              <Skeleton className="h-20 w-20 rounded-md sm:h-24 sm:w-24" />
              <Skeleton className="h-4 w-16" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <>
      <h1 className="mb-2 text-center text-3xl font-bold text-foreground sm:text-4xl">
        {t('whoIsWatching')}
      </h1>
      <p className="mb-8 text-center text-sm text-muted-foreground">
        {t('selectProfileToContinue')}
      </p>
      <div className="flex flex-wrap justify-center gap-4">
        {profiles.map((profile) => (
          <ProfileCard
            key={profile.id}
            profile={profile}
            onClick={() => handleSelectProfile(profile)}
          />
        ))}
        {profiles.length < MAX_PROFILES && (
          <button
            type="button"
            onClick={() => setIsModalOpen(true)}
            className={cn(
              'group flex flex-col items-center gap-2 rounded-lg p-4',
              'transition-transform hover:scale-105 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none'
            )}
          >
            <div className="flex h-20 w-20 items-center justify-center rounded-md border-2 border-dashed border-muted-foreground/40 transition-colors group-hover:border-foreground sm:h-24 sm:w-24">
              <Plus className="h-10 w-10 text-muted-foreground group-hover:text-foreground sm:h-12 sm:w-12" />
            </div>
            <span className="text-sm text-muted-foreground group-hover:text-foreground">
              {t('addProfile')}
            </span>
          </button>
        )}
      </div>
      {profiles.length >= MAX_PROFILES && (
        <p className="mt-4 text-center text-xs text-muted-foreground">
          {t('maxProfiles', { max: MAX_PROFILES })}
        </p>
      )}
      <CreateProfileModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
      />
      <Link
        href="/account/profiles"
        className="mt-8 text-sm text-muted-foreground underline-offset-4 hover:text-foreground hover:underline focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
      >
        {t('manageProfiles')}
      </Link>
    </>
  );
}
