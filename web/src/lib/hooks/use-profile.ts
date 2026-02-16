'use client';

import { useEffect } from 'react';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

import { getProfiles, createProfile } from '@/lib/api/auth';
import { useAuthStore } from '@/lib/stores/auth-store';
import type { CreateProfileRequest, Profile } from '@/lib/types/auth';

export function useProfile() {
  const queryClient = useQueryClient();
  const user = useAuthStore((s) => s.user);
  const activeProfile = useAuthStore((s) => s.activeProfile);
  const setActiveProfile = useAuthStore((s) => s.setActiveProfile);
  const setProfiles = useAuthStore((s) => s.setProfiles);

  const profilesQuery = useQuery({
    queryKey: ['auth', 'profiles', user?.id],
    queryFn: () => getProfiles(user?.id ?? ''),
    enabled: !!user?.id,
    staleTime: 5 * 60_000,
  });

  useEffect(() => {
    if (profilesQuery.data) {
      setProfiles(profilesQuery.data);
    }
  }, [profilesQuery.data, setProfiles]);

  const createProfileMutation = useMutation({
    mutationFn: (data: Omit<CreateProfileRequest, 'userId'>) =>
      createProfile({ ...data, userId: user?.id ?? '' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['auth', 'profiles'] });
    },
  });

  const switchProfile = (profile: Profile) => {
    setActiveProfile(profile);
  };

  return {
    profiles: profilesQuery.data ?? [],
    activeProfile,
    isLoading: profilesQuery.isLoading,
    switchProfile,
    createProfile: createProfileMutation.mutateAsync,
    isCreating: createProfileMutation.isPending,
  };
}
