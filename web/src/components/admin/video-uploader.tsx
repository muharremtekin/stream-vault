'use client';

import { useState, useRef, useCallback } from 'react';
import { useTranslations } from 'next-intl';
import { toast } from 'sonner';
import { Upload, FileVideo, CheckCircle, XCircle } from 'lucide-react';
import Link from 'next/link';

import { Button } from '@/components/ui/button';
import { ProgressBar } from '@/components/ui/progress-bar';
import { uploadVideo } from '@/lib/api/streaming';
import { extractErrorMessage } from '@/lib/utils/error';
import { cn } from '@/lib/utils/cn';
import { formatFileSize } from '@/lib/utils/format';

const ACCEPTED_TYPES = ['video/mp4', 'video/x-matroska', 'video/x-msvideo'];
const MAX_SIZE_BYTES = 10 * 1024 * 1024 * 1024; // 10GB
const ACCEPTED_EXTENSIONS = '.mp4,.mkv,.avi';

type UploadState = 'idle' | 'selected' | 'uploading' | 'complete' | 'error';

interface VideoUploaderProps {
  contentId: string;
}

export function VideoUploader({ contentId }: VideoUploaderProps) {
  const t = useTranslations('admin.upload');
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [state, setState] = useState<UploadState>('idle');
  const [file, setFile] = useState<File | null>(null);
  const [progress, setProgress] = useState(0);
  const [isDragging, setIsDragging] = useState(false);

  const handleFile = useCallback((f: File) => {
    if (!ACCEPTED_TYPES.includes(f.type) && !f.name.match(/\.(mp4|mkv|avi)$/i)) {
      toast.error(t('invalidFormat'));
      return;
    }
    if (f.size > MAX_SIZE_BYTES) {
      toast.error(t('tooLarge'));
      return;
    }
    setFile(f);
    setState('selected');
    setProgress(0);
  }, [t]);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const dropped = e.dataTransfer.files[0];
    if (dropped) handleFile(dropped);
  }, [handleFile]);

  const handleUpload = async () => {
    if (!file) return;
    setState('uploading');
    setProgress(0);
    try {
      const formData = new FormData();
      formData.append('content_id', contentId);
      formData.append('file', file);
      await uploadVideo(formData, (pct) => setProgress(pct));
      setState('complete');
      toast.success(t('successMessage'));
    } catch (err) {
      setState('error');
      toast.error(extractErrorMessage(err));
    }
  };

  const reset = () => {
    setFile(null);
    setState('idle');
    setProgress(0);
    if (fileInputRef.current) fileInputRef.current.value = '';
  };

  return (
    <div className="space-y-4">
      {(state === 'idle' || state === 'selected') && (
        <div
          onDragOver={(e) => { e.preventDefault(); setIsDragging(true); }}
          onDragLeave={() => setIsDragging(false)}
          onDrop={handleDrop}
          onClick={() => fileInputRef.current?.click()}
          className={cn(
            'flex cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed p-8 transition-colors',
            isDragging ? 'border-primary bg-primary/5' : 'border-border hover:border-muted-foreground'
          )}
        >
          <Upload className="mb-3 h-10 w-10 text-muted-foreground" />
          <p className="text-sm font-medium text-foreground">{t('dragDrop')}</p>
          <p className="mt-1 text-xs text-muted-foreground">{t('supportedFormats')}</p>
          <p className="text-xs text-muted-foreground">{t('maxSize')}</p>
          <input
            ref={fileInputRef}
            type="file"
            accept={ACCEPTED_EXTENSIONS}
            onChange={(e) => { const f = e.target.files?.[0]; if (f) handleFile(f); }}
            className="hidden"
          />
        </div>
      )}

      {state === 'selected' && file && (
        <div className="flex items-center justify-between rounded-lg border border-border p-4">
          <div className="flex items-center gap-3">
            <FileVideo className="h-5 w-5 text-muted-foreground" />
            <div>
              <p className="text-sm font-medium text-foreground">{file.name}</p>
              <p className="text-xs text-muted-foreground">{formatFileSize(file.size)}</p>
            </div>
          </div>
          <div className="flex gap-2">
            <Button size="sm" onClick={handleUpload}>{t('uploadButton')}</Button>
            <Button size="sm" variant="ghost" onClick={reset}>{t('cancel')}</Button>
          </div>
        </div>
      )}

      {state === 'uploading' && (
        <div className="rounded-lg border border-border p-4">
          <div className="mb-2 flex items-center justify-between text-sm">
            <span className="text-foreground">{t('uploading')}</span>
            <span className="text-muted-foreground">{progress}%</span>
          </div>
          <ProgressBar value={progress} variant="primary" size="md" />
          {file && (
            <p className="mt-2 text-xs text-muted-foreground">
              {formatFileSize(file.size * progress / 100)} / {formatFileSize(file.size)}
            </p>
          )}
        </div>
      )}

      {state === 'complete' && (
        <div className="flex items-center gap-3 rounded-lg border border-success/30 bg-success/5 p-4">
          <CheckCircle className="h-5 w-5 text-success" />
          <div className="flex-1">
            <p className="text-sm font-medium text-foreground">{t('complete')}</p>
            <Link href="/admin/encoding" className="text-xs text-primary hover:underline">
              {t('viewEncodingJobs')}
            </Link>
          </div>
          <Button size="sm" variant="ghost" onClick={reset}>{t('uploadAnother')}</Button>
        </div>
      )}

      {state === 'error' && (
        <div className="flex items-center gap-3 rounded-lg border border-destructive/30 bg-destructive/5 p-4">
          <XCircle className="h-5 w-5 text-destructive" />
          <div className="flex-1">
            <p className="text-sm font-medium text-foreground">{t('error')}</p>
          </div>
          <Button size="sm" variant="ghost" onClick={reset}>{t('tryAgain')}</Button>
        </div>
      )}
    </div>
  );
}
