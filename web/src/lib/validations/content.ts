import { z } from 'zod';

const castMemberSchema = z.object({
  id: z.string().optional(),
  name: z.string().min(1, 'Cast member name is required').max(100),
  role: z.string().min(1, 'Cast member role is required').max(100),
  photoUrl: z.string().url().nullable().optional(),
});

const MATURITY_RATINGS = ['G', 'PG', 'PG-13', 'R', 'NC-17', 'TV-Y', 'TV-G', 'TV-PG', 'TV-14', 'TV-MA'] as const;

const currentYear = new Date().getFullYear();

const baseContentSchema = z.object({
  title: z.string().min(1, 'Title is required').max(200, 'Title must be 200 characters or less'),
  originalTitle: z.string().max(200).optional().or(z.literal('')),
  description: z.string().min(1, 'Description is required').max(2000, 'Description must be 2000 characters or less'),
  releaseYear: z.number()
    .int()
    .min(1888, 'Year must be 1888 or later')
    .max(currentYear + 5, `Year must be ${currentYear + 5} or earlier`),
  maturityRating: z.string().min(1, 'Maturity rating is required'),
  genres: z.array(z.string()).min(1, 'At least one genre is required'),
  cast: z.array(castMemberSchema),
  thumbnailUrl: z.string().min(1, 'Thumbnail URL is required').url('Must be a valid URL'),
  bannerUrl: z.string().min(1, 'Banner URL is required').url('Must be a valid URL'),
  trailerUrl: z.string().url('Must be a valid URL').optional().or(z.literal('')),
  tags: z.array(z.string()),
});

export const movieSchema = baseContentSchema.extend({
  durationMinutes: z.number()
    .int()
    .min(1, 'Duration must be at least 1 minute')
    .max(600, 'Duration must be 600 minutes or less'),
  director: z.string().min(1, 'Director is required').max(100, 'Director must be 100 characters or less'),
});

export const seriesSchema = baseContentSchema.extend({
  creator: z.string().min(1, 'Creator is required').max(100, 'Creator must be 100 characters or less'),
});

export type MovieFormData = z.infer<typeof movieSchema>;
export type SeriesFormData = z.infer<typeof seriesSchema>;

/** Combined schema for the form — all fields present, type-specific ones optional */
export const contentFormSchema = baseContentSchema.extend({
  durationMinutes: z.number().int().min(1).max(600).optional(),
  director: z.string().max(100).optional().or(z.literal('')),
  creator: z.string().max(100).optional().or(z.literal('')),
});

export type ContentFormData = z.infer<typeof contentFormSchema>;

export const MATURITY_RATING_OPTIONS = MATURITY_RATINGS;
