'use client';

import { useEffect, useRef, useState } from 'react';

import Hls from 'hls.js';

interface UseHlsPlayerOptions {
  manifestUrl: string | null;
  isEnabled: boolean;
}

interface UseHlsPlayerResult {
  isReady: boolean;
  error: string | null;
  hlsRef: React.RefObject<Hls | null>;
}

export function useHlsPlayer(
  videoRef: React.RefObject<HTMLVideoElement | null>,
  options: UseHlsPlayerOptions,
): UseHlsPlayerResult {
  const { manifestUrl, isEnabled } = options;
  const hlsRef = useRef<Hls | null>(null);
  const [isReady, setIsReady] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isEnabled || !manifestUrl) return;

    const video = videoRef.current;
    if (!video) return;

    setIsReady(false);
    setError(null);

    // Safari: native HLS support — no hls.js needed
    if (!Hls.isSupported()) {
      if (video.canPlayType('application/vnd.apple.mpegurl')) {
        video.src = manifestUrl;
        const onLoaded = () => setIsReady(true);
        video.addEventListener('loadedmetadata', onLoaded);
        return () => video.removeEventListener('loadedmetadata', onLoaded);
      }
      setError('unsupported');
      return;
    }

    const hls = new Hls({ startLevel: -1, enableWorker: true });
    hlsRef.current = hls;

    hls.loadSource(manifestUrl);
    hls.attachMedia(video);

    hls.on(Hls.Events.MANIFEST_PARSED, () => {
      setIsReady(true);
    });

    hls.on(Hls.Events.ERROR, (_event, data) => {
      if (!data.fatal) return;

      if (data.type === Hls.ErrorTypes.NETWORK_ERROR) {
        hls.startLoad();
      } else if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
        hls.recoverMediaError();
      } else {
        setError('fatal');
        hls.destroy();
        hlsRef.current = null;
      }
    });

    return () => {
      hls.destroy();
      hlsRef.current = null;
    };
  }, [manifestUrl, isEnabled, videoRef]);

  return { isReady, error, hlsRef };
}
