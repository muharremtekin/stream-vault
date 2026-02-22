'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useTranslations } from 'next-intl';
import { toast } from 'sonner';
import { Film, Tv } from 'lucide-react';

import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { TagInput } from '@/components/admin/tag-input';
import { CastInput } from '@/components/admin/cast-input';
import { useCreateMovie, useCreateSeries } from '@/lib/hooks/use-admin';
import { extractErrorMessage } from '@/lib/utils/error';
import { cn } from '@/lib/utils/cn';
import { contentFormSchema, MATURITY_RATING_OPTIONS, MATURITY_RATING_LABELS } from '@/lib/validations/content';

import type { ContentFormData } from '@/lib/validations/content';

type ContentType = 'movie' | 'series';

const DEFAULTS: ContentFormData = {
  title: '', originalTitle: '', description: '',
  releaseYear: new Date().getFullYear(), maturityRating: '',
  genres: [], cast: [], thumbnailUrl: '', bannerUrl: '',
  trailerUrl: '', tags: [], durationMinutes: 90, director: '', creator: '',
};

export function ContentForm() {
  const t = useTranslations('admin.contentForm');
  const router = useRouter();
  const [contentType, setContentType] = useState<ContentType>('movie');
  const createMovie = useCreateMovie();
  const createSeries = useCreateSeries();
  const isMovie = contentType === 'movie';

  const { register, handleSubmit, control, reset, formState: { errors, isSubmitting } } =
    useForm<ContentFormData>({ resolver: zodResolver(contentFormSchema), defaultValues: DEFAULTS });

  const switchType = (type: ContentType) => { setContentType(type); reset(DEFAULTS); };

  const onSubmit = async (data: ContentFormData) => {
    try {
      let result: { id: string };
      if (isMovie) {
        if (!data.director) { toast.error(t('directorRequired')); return; }
        if (!data.durationMinutes) { toast.error(t('durationRequired')); return; }
        result = await createMovie.mutateAsync({
          ...data, director: data.director, durationMinutes: data.durationMinutes,
        });
      } else {
        if (!data.creator) { toast.error(t('creatorRequired')); return; }
        result = await createSeries.mutateAsync({ ...data, creator: data.creator });
      }
      toast.success(t('success'));
      router.push(`/admin/content/${result.id}`);
    } catch (err) {
      toast.error(extractErrorMessage(err));
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} noValidate className="space-y-6">
      <TypeToggle value={contentType} onChange={switchType} t={t} />
      <div className="grid gap-4 sm:grid-cols-2">
        <Input label={t('title')} error={errors.title?.message} {...register('title')} />
        <Input label={t('originalTitle')} {...register('originalTitle')} />
      </div>
      <div>
        <label className="mb-1.5 block text-sm font-medium text-foreground">{t('description')}</label>
        <textarea {...register('description')} rows={4} className={cn(
          'w-full rounded-md border bg-input px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground transition-colors',
          'focus-visible:ring-2 focus-visible:ring-ring focus-visible:border-ring focus-visible:outline-none',
          errors.description ? 'border-destructive' : 'border-input-border'
        )} />
        {errors.description && <p className="mt-1.5 text-xs text-destructive">{errors.description.message}</p>}
      </div>
      <div className="grid gap-4 sm:grid-cols-3">
        <Input label={t('releaseYear')} type="number" error={errors.releaseYear?.message} {...register('releaseYear', { valueAsNumber: true })} />
        {isMovie && (
          <Input label={t('durationMinutes')} type="number" error={errors.durationMinutes?.message} {...register('durationMinutes', { valueAsNumber: true })} />
        )}
        <div>
          <label className="mb-1.5 block text-sm font-medium text-foreground">{t('maturityRating')}</label>
          <select {...register('maturityRating')} className={cn(
            'h-10 w-full rounded-md border bg-input px-3 text-sm text-foreground transition-colors',
            'focus-visible:ring-2 focus-visible:ring-ring focus-visible:border-ring focus-visible:outline-none',
            errors.maturityRating ? 'border-destructive' : 'border-input-border'
          )}>
            <option value="">{t('selectRating')}</option>
            {MATURITY_RATING_OPTIONS.map((r) => <option key={r} value={r}>{MATURITY_RATING_LABELS[r] ?? r}</option>)}
          </select>
          {errors.maturityRating && <p className="mt-1.5 text-xs text-destructive">{errors.maturityRating.message}</p>}
        </div>
      </div>
      {isMovie
        ? <Input label={t('director')} error={errors.director?.message} {...register('director')} />
        : <Input label={t('creator')} error={errors.creator?.message} {...register('creator')} />}
      <Controller name="genres" control={control} render={({ field }) => (
        <TagInput label={t('genres')} value={field.value ?? []} onChange={field.onChange}
          placeholder={t('genresPlaceholder')} error={errors.genres?.message} />
      )} />
      <Controller name="tags" control={control} render={({ field }) => (
        <TagInput label={t('tags')} value={field.value ?? []} onChange={field.onChange} placeholder={t('tagsPlaceholder')} />
      )} />
      <Controller name="cast" control={control} render={({ field }) => (
        <CastInput value={field.value ?? []} onChange={field.onChange} error={errors.cast?.message} />
      )} />
      <div className="grid gap-4 sm:grid-cols-2">
        <Input label={t('thumbnailUrl')} error={errors.thumbnailUrl?.message} {...register('thumbnailUrl')} />
        <Input label={t('bannerUrl')} error={errors.bannerUrl?.message} {...register('bannerUrl')} />
      </div>
      <Input label={t('trailerUrl')} {...register('trailerUrl')} />
      <div className="flex gap-3 pt-2">
        <Button type="submit" isLoading={isSubmitting}>{t('save')}</Button>
        <Button type="button" variant="ghost" onClick={() => router.push('/admin/content')}>{t('cancel')}</Button>
      </div>
    </form>
  );
}

interface TypeToggleProps {
  value: ContentType;
  onChange: (type: ContentType) => void;
  t: (key: string) => string;
}

function TypeToggle({ value, onChange, t }: TypeToggleProps) {
  return (
    <div className="flex gap-2">
      <button type="button" onClick={() => onChange('movie')} className={cn(
        'flex items-center gap-2 rounded-lg border px-4 py-2 text-sm font-medium transition-colors',
        value === 'movie' ? 'border-primary bg-primary/10 text-primary' : 'border-border text-muted-foreground hover:bg-muted/50'
      )}>
        <Film className="h-4 w-4" /> {t('movie')}
      </button>
      <button type="button" onClick={() => onChange('series')} className={cn(
        'flex items-center gap-2 rounded-lg border px-4 py-2 text-sm font-medium transition-colors',
        value === 'series' ? 'border-primary bg-primary/10 text-primary' : 'border-border text-muted-foreground hover:bg-muted/50'
      )}>
        <Tv className="h-4 w-4" /> {t('series')}
      </button>
    </div>
  );
}
