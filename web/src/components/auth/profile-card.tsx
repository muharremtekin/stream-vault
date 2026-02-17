'use client';

import {
  Smile, Cat, Dog, Bird, Fish,
  Rabbit, Star, Heart, Ghost, Rocket,
} from 'lucide-react';

import { cn } from '@/lib/utils/cn';
import { getProfileIcon } from '@/lib/utils/profile-icons';
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

interface ProfileCardProps {
  profile: Profile;
  onClick: () => void;
}

export function ProfileCard({ profile, onClick }: ProfileCardProps) {
  const iconMeta = getProfileIcon(profile.icon);
  const IconComponent = ICON_MAP[iconMeta.id] ?? Smile;

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'group flex flex-col items-center gap-2 rounded-lg p-4',
        'transition-transform hover:scale-105 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none'
      )}
    >
      <div
        className="flex h-20 w-20 items-center justify-center rounded-md transition-colors group-hover:ring-2 group-hover:ring-foreground sm:h-24 sm:w-24"
        style={{ backgroundColor: `${iconMeta.color}20` }}
      >
        <IconComponent className="h-10 w-10 sm:h-12 sm:w-12" style={{ color: iconMeta.color }} />
      </div>
      <span className="text-sm text-muted-foreground group-hover:text-foreground">
        {profile.name}
      </span>
    </button>
  );
}
