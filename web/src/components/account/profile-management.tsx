'use client';

import { useState } from 'react';

import {
  Smile, Cat, Dog, Bird, Fish,
  Rabbit, Star, Heart, Ghost, Rocket,
  Pencil, Trash2, Plus, Baby,
} from 'lucide-react';
import { useTranslations } from 'next-intl';
import { toast } from 'sonner';

import { useProfile } from '@/lib/hooks/use-profile';
import { cn } from '@/lib/utils/cn';
import { extractErrorMessage } from '@/lib/utils/error';
import { getProfileIcon } from '@/lib/utils/profile-icons';
import { MAX_PROFILES } from '@/lib/utils/constants';
import { Button } from '@/components/ui/button';
import { Modal } from '@/components/ui/modal';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { CreateProfileModal } from '@/components/auth/create-profile-modal';
import { EditProfileModal } from '@/components/account/edit-profile-modal';
import type { Profile } from '@/lib/types/auth';

const ICON_MAP: Record<string, React.ComponentType<{ className?: string; style?: React.CSSProperties }>> = {
  smile: Smile, cat: Cat, dog: Dog, bird: Bird, fish: Fish,
  rabbit: Rabbit, star: Star, heart: Heart, ghost: Ghost, rocket: Rocket,
};

export function ProfileManagement() {
  const t = useTranslations('account.profiles');
  const tAuth = useTranslations('auth');
  const tCommon = useTranslations('common');
  const { profiles, activeProfile, isLoading, deleteProfile, isDeleting } = useProfile();

  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingProfile, setEditingProfile] = useState<Profile | null>(null);
  const [deletingProfile, setDeletingProfile] = useState<Profile | null>(null);

  const handleDelete = async () => {
    if (!deletingProfile) return;
    try {
      await deleteProfile(deletingProfile.id);
      toast.success(t('profileDeleted'));
      setDeletingProfile(null);
    } catch (error: unknown) {
      toast.error(extractErrorMessage(error));
    }
  };

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={`profile-skel-${i}`} className="rounded-lg border border-border bg-card p-6">
            <div className="flex items-center gap-4">
              <Skeleton className="h-16 w-16 rounded-md" />
              <div className="flex-1 space-y-2">
                <Skeleton variant="text" className="h-5 w-24" />
                <Skeleton variant="text" className="h-4 w-16" />
              </div>
            </div>
          </div>
        ))}
      </div>
    );
  }

  return (
    <>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {profiles.map((profile) => {
          const iconMeta = getProfileIcon(profile.icon);
          const IconComponent = ICON_MAP[iconMeta.id] ?? Smile;
          const isActive = activeProfile?.id === profile.id;

          return (
            <div
              key={profile.id}
              className={cn(
                'rounded-lg border bg-card p-5 transition-colors',
                isActive ? 'border-primary' : 'border-border'
              )}
            >
              <div className="flex items-center gap-4">
                <div
                  className="flex h-14 w-14 items-center justify-center rounded-md"
                  style={{ backgroundColor: `${iconMeta.color}20` }}
                >
                  <IconComponent
                    className="h-8 w-8"
                    style={{ color: iconMeta.color }}
                  />
                </div>
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <p className="font-medium text-foreground">{profile.name}</p>
                    {isActive && (
                      <Badge variant="success">{t('activeProfile')}</Badge>
                    )}
                  </div>
                  {profile.isKids && (
                    <div className="mt-1 flex items-center gap-1 text-xs text-muted-foreground">
                      <Baby className="h-3 w-3" />
                      <span>{tAuth('kidsProfile')}</span>
                    </div>
                  )}
                </div>
              </div>
              <div className="mt-4 flex gap-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setEditingProfile(profile)}
                  aria-label={`${t('editProfile')} ${profile.name}`}
                >
                  <Pencil className="mr-1.5 h-3.5 w-3.5" />
                  {tCommon('edit')}
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setDeletingProfile(profile)}
                  aria-label={`${t('deleteProfile')} ${profile.name}`}
                >
                  <Trash2 className="mr-1.5 h-3.5 w-3.5" />
                  {tCommon('delete')}
                </Button>
              </div>
            </div>
          );
        })}

        {/* Add Profile Card */}
        {profiles.length < MAX_PROFILES && (
          <button
            type="button"
            onClick={() => setIsCreateOpen(true)}
            className={cn(
              'flex flex-col items-center justify-center gap-3 rounded-lg border-2 border-dashed border-muted-foreground/30 p-6',
              'transition-colors hover:border-foreground/50 hover:bg-muted/30',
              'focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none'
            )}
          >
            <div className="flex h-14 w-14 items-center justify-center rounded-full bg-muted">
              <Plus className="h-7 w-7 text-muted-foreground" />
            </div>
            <span className="text-sm font-medium text-muted-foreground">
              {tAuth('addProfile')}
            </span>
          </button>
        )}
      </div>

      {profiles.length >= MAX_PROFILES && (
        <p className="mt-4 text-center text-xs text-muted-foreground">
          {tAuth('maxProfiles', { max: MAX_PROFILES })}
        </p>
      )}

      {/* Create Profile Modal */}
      <CreateProfileModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
      />

      {/* Edit Profile Modal */}
      {editingProfile && (
        <EditProfileModal
          isOpen={!!editingProfile}
          onClose={() => setEditingProfile(null)}
          profile={editingProfile}
        />
      )}

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={!!deletingProfile}
        onClose={() => setDeletingProfile(null)}
        title={t('deleteConfirmTitle')}
        size="sm"
      >
        <p className="text-sm text-muted-foreground">
          {t('deleteConfirmMessage', { name: deletingProfile?.name ?? '' })}
        </p>
        <div className="mt-6 flex justify-end gap-3">
          <Button
            variant="secondary"
            onClick={() => setDeletingProfile(null)}
          >
            {tCommon('cancel')}
          </Button>
          <Button
            variant="danger"
            onClick={handleDelete}
            isLoading={isDeleting}
          >
            {tCommon('delete')}
          </Button>
        </div>
      </Modal>
    </>
  );
}
