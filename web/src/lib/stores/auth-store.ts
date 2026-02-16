'use client';

import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

import type { User, Profile } from '@/lib/types/auth';
import type { SubscriptionTier } from '@/lib/types/common';

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  activeProfile: Profile | null;
  profiles: Profile[];
  subscriptionTier: SubscriptionTier | null;

  setAuth: (
    user: User,
    accessToken: string,
    refreshToken: string
  ) => void;
  setTokens: (accessToken: string, refreshToken: string) => void;
  setUser: (user: User) => void;
  setActiveProfile: (profile: Profile) => void;
  setProfiles: (profiles: Profile[]) => void;
  setSubscriptionTier: (tier: SubscriptionTier) => void;
  logout: () => void;
}

const initialState = {
  user: null,
  accessToken: null,
  refreshToken: null,
  activeProfile: null,
  profiles: [],
  subscriptionTier: null,
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      ...initialState,

      setAuth: (user, accessToken, refreshToken) =>
        set({ user, accessToken, refreshToken }),

      setTokens: (accessToken, refreshToken) =>
        set({ accessToken, refreshToken }),

      setUser: (user) => set({ user }),

      setActiveProfile: (profile) => set({ activeProfile: profile }),

      setProfiles: (profiles) => set({ profiles }),

      setSubscriptionTier: (tier) => set({ subscriptionTier: tier }),

      logout: () => set(initialState),
    }),
    {
      name: 'streamvault-auth',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        activeProfile: state.activeProfile,
        profiles: state.profiles,
        subscriptionTier: state.subscriptionTier,
      }),
    }
  )
);
