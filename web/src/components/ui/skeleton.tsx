import { cn } from '@/lib/utils/cn';

const SKELETON_VARIANTS = {
  text: 'h-4 w-full rounded',
  card: 'h-64 w-44 rounded-md',
  circle: 'h-10 w-10 rounded-full',
  custom: '',
} as const;

type SkeletonVariant = keyof typeof SKELETON_VARIANTS;

interface SkeletonProps {
  variant?: SkeletonVariant;
  className?: string;
}

export function Skeleton({ variant = 'text', className }: SkeletonProps) {
  return (
    <div
      className={cn('animate-pulse bg-muted', SKELETON_VARIANTS[variant], className)}
    />
  );
}
