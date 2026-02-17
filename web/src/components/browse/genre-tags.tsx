'use client';

import { useRouter } from 'next/navigation';

import { cn } from '@/lib/utils/cn';

interface GenreTagsProps {
  genres: string[];
  className?: string;
}

function slugifyGenre(genre: string): string {
  return genre.toLowerCase().replace(/\s+/g, '-');
}

export default function GenreTags({ genres, className }: GenreTagsProps) {
  const router = useRouter();

  if (genres.length === 0) return null;

  return (
    <div className={cn('flex flex-wrap gap-1', className)}>
      {genres.map((genre) => (
        <button
          key={genre}
          type="button"
          onClick={() => router.push(`/genre/${slugifyGenre(genre)}`)}
          className="rounded px-2 py-0.5 text-xs bg-white/10 hover:bg-white/20 transition-colors cursor-pointer"
        >
          {genre}
        </button>
      ))}
    </div>
  );
}
