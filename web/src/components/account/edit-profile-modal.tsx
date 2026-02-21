'use client';

import { useEffect } from 'react';

import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import {
  Smile, Cat, Dog, Bird, Fish,
  Rabbit, Star, Heart, Ghost, Rocket,
} from 'lucide-react';
import { useTranslations } from 'next-intl';
import { toast } from 'sonner';

import { useProfile } from '@/lib/hooks/use-profile';
import { cn } from '@/lib/utils/cn';
import { extractErrorMessage } from '@/lib/utils/error';
import { PROFILE_ICONS } from '@/lib/utils/profile-icons';
import { createProfileSchema } from '@/lib/validations/auth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Modal } from '@/components/ui/modal';
import type { CreateProfileFormData } from '@/lib/validations/auth';
import type { Profile } from '@/lib/types/auth';

const ICON_MAP: Record<string, React.ComponentType<{ className?: string; style?: React.CSSProperties }>> = {
  smile: Smile,
  cat: Cat,
  dog: Dog,
  bird: Bird,
  fish: Fish,
  rabbit: Rabbit,
  star: Star,
  heart: Heart,
  ghost: Ghost,
  rocket: Rocket,
};

interface EditProfileModalProps {
  isOpen: boolean;
  onClose: () => void;
  profile: Profile;
}

export function EditProfileModal({ isOpen, onClose, profile }: EditProfileModalProps) {
  const t = useTranslations('account.profiles');
  const tAuth = useTranslations('auth');
  const { updateProfile, isUpdating } = useProfile();

  const {
    register,
    handleSubmit,
    control,
    reset,
    formState: { errors },
  } = useForm<CreateProfileFormData>({
    resolver: zodResolver(createProfileSchema),
    defaultValues: {
      name: profile.name,
      icon: profile.icon,
      isKids: profile.isKids,
    },
  });

  useEffect(() => {
    if (isOpen) {
      reset({
        name: profile.name,
        icon: profile.icon,
        isKids: profile.isKids,
      });
    }
  }, [isOpen, profile, reset]);

  const onSubmit = async (data: CreateProfileFormData) => {
    try {
      await updateProfile({ id: profile.id, data });
      toast.success(t('profileUpdated'));
      onClose();
    } catch (error: unknown) {
      toast.error(extractErrorMessage(error));
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={t('editProfile')} size="md">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
        <div>
          <p className="mb-3 text-sm font-medium text-foreground">
            {tAuth('selectAvatar')}
          </p>
          <Controller
            name="icon"
            control={control}
            render={({ field }) => (
              <div className="grid grid-cols-5 gap-2">
                {PROFILE_ICONS.map((iconMeta) => {
                  const IconComp = ICON_MAP[iconMeta.id] ?? Smile;
                  const isSelected = field.value === iconMeta.id;
                  return (
                    <button
                      key={iconMeta.id}
                      type="button"
                      onClick={() => field.onChange(iconMeta.id)}
                      className={cn(
                        'flex h-14 w-14 items-center justify-center rounded-md transition-all',
                        'focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none',
                        isSelected
                          ? 'ring-2 ring-primary scale-110'
                          : 'hover:scale-105 hover:ring-1 hover:ring-muted-foreground'
                      )}
                      style={{ backgroundColor: `${iconMeta.color}20` }}
                      aria-label={iconMeta.id}
                      aria-pressed={isSelected}
                    >
                      <IconComp
                        className="h-7 w-7"
                        style={{ color: iconMeta.color }}
                      />
                    </button>
                  );
                })}
              </div>
            )}
          />
          {errors.icon?.message && (
            <p className="mt-1.5 text-xs text-destructive">
              {errors.icon.message}
            </p>
          )}
        </div>
        <Input
          label={tAuth('profileName')}
          placeholder={tAuth('profileName')}
          error={errors.name?.message}
          {...register('name')}
        />
        <label className="flex items-center gap-3">
          <input
            type="checkbox"
            className="h-4 w-4 rounded border-border bg-input accent-primary"
            {...register('isKids')}
          />
          <span className="text-sm text-foreground">{tAuth('kidsProfile')}</span>
        </label>
        <Button type="submit" isLoading={isUpdating} className="w-full">
          {t('editProfile')}
        </Button>
      </form>
    </Modal>
  );
}
