'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { toast } from 'sonner';
import { Plus, Film, Clock, ChevronDown, ChevronUp, Upload, CheckCircle, Loader2, AlertCircle } from 'lucide-react';

import { useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { VideoUploader } from '@/components/admin/video-uploader';
import { useAddEpisode } from '@/lib/hooks/use-admin';
import { extractErrorMessage } from '@/lib/utils/error';
import { cn } from '@/lib/utils/cn';
import { buildEpisodeContentId } from '@/lib/utils/episode-content-id';
import type { Series, Episode } from '@/lib/types/catalog';

function VideoStatusBadge({ status }: { status: string }) {
  const t = useTranslations('admin.episodes');
  switch (status) {
    case 'Ready':
      return (
        <Badge size="sm" variant="success" className="gap-1">
          <CheckCircle className="h-3 w-3" />
          {t('statusReady')}
        </Badge>
      );
    case 'Encoding':
    case 'Queued':
    case 'Uploading':
      return (
        <Badge size="sm" variant="warning" className="gap-1">
          <Loader2 className="h-3 w-3 animate-spin" />
          {t('statusProcessing')}
        </Badge>
      );
    case 'Error':
      return (
        <Badge size="sm" variant="error" className="gap-1">
          <AlertCircle className="h-3 w-3" />
          {t('statusError')}
        </Badge>
      );
    default:
      return (
        <Badge size="sm" variant="default" className="gap-1 opacity-60">
          <Upload className="h-3 w-3" />
          {t('statusNoVideo')}
        </Badge>
      );
  }
}

interface EpisodeManagerProps {
  series: Series;
}

export function EpisodeManager({ series }: EpisodeManagerProps) {
  const t = useTranslations('admin.episodes');
  const queryClient = useQueryClient();
  const [expandedSeason, setExpandedSeason] = useState<number | null>(
    series.seasons.length > 0 ? series.seasons[0].seasonNumber : null
  );
  const [showAddForm, setShowAddForm] = useState(false);
  const [addSeasonNumber, setAddSeasonNumber] = useState(
    series.seasons.length > 0 ? series.seasons[0].seasonNumber : 1
  );
  const [uploadingEpisode, setUploadingEpisode] = useState<string | null>(null);

  const toggleSeason = (seasonNumber: number) => {
    setExpandedSeason((prev) => (prev === seasonNumber ? null : seasonNumber));
  };

  const toggleUpload = (seasonNumber: number, episode: Episode) => {
    const key = `${seasonNumber}_${episode.episodeNumber}`;
    setUploadingEpisode((prev) => (prev === key ? null : key));
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-foreground">{t('title')}</h2>
        <Button
          size="sm"
          onClick={() => setShowAddForm((prev) => !prev)}
          className="gap-1.5"
        >
          <Plus className="h-3.5 w-3.5" />
          {t('addEpisode')}
        </Button>
      </div>

      {showAddForm && (
        <AddEpisodeForm
          seriesId={series.id}
          seasons={series.seasons}
          defaultSeason={addSeasonNumber}
          onSeasonChange={setAddSeasonNumber}
          onSuccess={() => setShowAddForm(false)}
        />
      )}

      {series.seasons.length === 0 ? (
        <div className="rounded-lg border border-border bg-card p-8 text-center">
          <Film className="mx-auto mb-3 h-10 w-10 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">{t('noSeasons')}</p>
          <p className="mt-1 text-xs text-muted-foreground">{t('noSeasonsHint')}</p>
        </div>
      ) : (
        <div className="space-y-2">
          {series.seasons.map((season) => {
            const isExpanded = expandedSeason === season.seasonNumber;
            return (
              <div
                key={season.seasonNumber}
                className="rounded-lg border border-border overflow-hidden"
              >
                <button
                  type="button"
                  onClick={() => toggleSeason(season.seasonNumber)}
                  className={cn(
                    'flex w-full items-center justify-between p-4 text-left transition-colors',
                    'hover:bg-muted/50',
                    isExpanded && 'bg-muted/30'
                  )}
                >
                  <div className="flex items-center gap-3">
                    <span className="text-sm font-medium text-foreground">
                      {t('seasonLabel', { number: season.seasonNumber })}
                    </span>
                    {season.title && (
                      <span className="text-sm text-muted-foreground">
                        — {season.title}
                      </span>
                    )}
                    <Badge size="sm" variant="default">
                      {t('episodeCount', { count: season.episodeCount })}
                    </Badge>
                  </div>
                  {isExpanded ? (
                    <ChevronUp className="h-4 w-4 text-muted-foreground" />
                  ) : (
                    <ChevronDown className="h-4 w-4 text-muted-foreground" />
                  )}
                </button>

                {isExpanded && (
                  <div className="border-t border-border">
                    {season.episodes.length === 0 ? (
                      <div className="p-4 text-center">
                        <p className="text-sm text-muted-foreground">
                          {t('noEpisodes')}
                        </p>
                      </div>
                    ) : (
                      <div className="divide-y divide-border">
                        {season.episodes.map((episode) => {
                          const uploadKey = `${season.seasonNumber}_${episode.episodeNumber}`;
                          const isUploading = uploadingEpisode === uploadKey;
                          const contentId = buildEpisodeContentId(
                            series.id,
                            season.seasonNumber,
                            episode.episodeNumber
                          );

                          return (
                            <div key={episode.episodeNumber} className="space-y-0">
                              <div className="flex items-center gap-4 px-4 py-3">
                                <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-muted text-xs font-medium text-muted-foreground">
                                  {episode.episodeNumber}
                                </span>
                                <div className="min-w-0 flex-1">
                                  <p className="text-sm font-medium text-foreground truncate">
                                    {episode.title}
                                  </p>
                                  {episode.description && (
                                    <p className="text-xs text-muted-foreground line-clamp-1">
                                      {episode.description}
                                    </p>
                                  )}
                                </div>
                                <div className="flex items-center gap-2 shrink-0">
                                  <VideoStatusBadge status={episode.videoStatus} />
                                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                                    <Clock className="h-3 w-3" />
                                    {episode.durationFormatted || `${episode.durationMinutes}m`}
                                  </div>
                                  <Button
                                    size="sm"
                                    variant="ghost"
                                    onClick={() => toggleUpload(season.seasonNumber, episode)}
                                    className="gap-1 text-xs"
                                  >
                                    <Upload className="h-3 w-3" />
                                    {t('uploadVideo')}
                                  </Button>
                                </div>
                              </div>
                              {isUploading && (
                                <div className="px-4 pb-3">
                                  <VideoUploader
                                  contentId={contentId}
                                  onUploadComplete={() => {
                                    queryClient.invalidateQueries({ queryKey: ['admin', 'seriesDetail', series.id] });
                                  }}
                                />
                                </div>
                              )}
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

interface AddEpisodeFormProps {
  seriesId: string;
  seasons: Series['seasons'];
  defaultSeason: number;
  onSeasonChange: (season: number) => void;
  onSuccess: () => void;
}

function AddEpisodeForm({
  seriesId,
  seasons,
  defaultSeason,
  onSeasonChange,
  onSuccess,
}: AddEpisodeFormProps) {
  const t = useTranslations('admin.episodes');
  const addEpisodeMutation = useAddEpisode(seriesId);

  const [seasonNumber, setSeasonNumber] = useState(defaultSeason);
  const [episodeNumber, setEpisodeNumber] = useState('');
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [durationMinutes, setDurationMinutes] = useState('');
  const [thumbnailUrl, setThumbnailUrl] = useState('');

  const handleSeasonChange = (val: number) => {
    setSeasonNumber(val);
    onSeasonChange(val);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const epNum = parseInt(episodeNumber, 10);
    const dur = parseInt(durationMinutes, 10);

    if (!title.trim() || isNaN(epNum) || isNaN(dur)) {
      toast.error(t('validationError'));
      return;
    }

    try {
      await addEpisodeMutation.mutateAsync({
        seasonNumber,
        data: {
          episodeNumber: epNum,
          title: title.trim(),
          description: description.trim(),
          durationMinutes: dur,
          thumbnailUrl: thumbnailUrl.trim(),
        },
      });
      toast.success(t('addSuccess'));
      setEpisodeNumber('');
      setTitle('');
      setDescription('');
      setDurationMinutes('');
      setThumbnailUrl('');
      onSuccess();
    } catch (err) {
      toast.error(extractErrorMessage(err));
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="rounded-lg border border-primary/20 bg-primary/5 p-4 space-y-3"
    >
      <h3 className="text-sm font-medium text-foreground">{t('addEpisode')}</h3>

      <div className="grid gap-3 sm:grid-cols-3">
        <div>
          <label htmlFor="ep-season" className="mb-1 block text-xs font-medium text-muted-foreground">
            {t('season')}
          </label>
          <select
            id="ep-season"
            value={seasonNumber}
            onChange={(e) => handleSeasonChange(Number(e.target.value))}
            className="h-10 w-full rounded-md border border-input-border bg-input px-3 text-sm text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
          >
            {seasons.map((s) => (
              <option key={s.seasonNumber} value={s.seasonNumber}>
                {t('seasonLabel', { number: s.seasonNumber })}
              </option>
            ))}
            <option value={seasons.length > 0 ? Math.max(...seasons.map((s) => s.seasonNumber)) + 1 : 1}>
              {t('newSeason')}
            </option>
          </select>
        </div>
        <Input
          label={t('episodeNumber')}
          type="number"
          min={1}
          value={episodeNumber}
          onChange={(e) => setEpisodeNumber(e.target.value)}
          placeholder="1"
          required
        />
        <Input
          label={t('duration')}
          type="number"
          min={1}
          value={durationMinutes}
          onChange={(e) => setDurationMinutes(e.target.value)}
          placeholder="45"
          required
        />
      </div>

      <Input
        label={t('episodeTitle')}
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder={t('episodeTitlePlaceholder')}
        required
      />

      <div>
        <label htmlFor="ep-desc" className="mb-1 block text-xs font-medium text-muted-foreground">
          {t('description')}
        </label>
        <textarea
          id="ep-desc"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder={t('descriptionPlaceholder')}
          rows={2}
          className="w-full rounded-md border border-input-border bg-input px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none resize-none"
        />
      </div>

      <Input
        label={t('thumbnailUrl')}
        value={thumbnailUrl}
        onChange={(e) => setThumbnailUrl(e.target.value)}
        placeholder="https://..."
      />

      <div className="flex justify-end gap-2">
        <Button
          type="submit"
          size="sm"
          disabled={addEpisodeMutation.isPending}
        >
          {addEpisodeMutation.isPending ? t('adding') : t('add')}
        </Button>
      </div>
    </form>
  );
}
