const EPISODE_CONTENT_ID_PATTERN =
  /^(?<seriesId>[a-f0-9]{24})_s(?<seasonNumber>\d+)_e(?<episodeNumber>\d+)$/i;

/**
 * Builds the stable playback/progress identifier used for a series episode.
 *
 * @param {string} seriesId
 * @param {number} seasonNumber
 * @param {number} episodeNumber
 */
export function buildEpisodeContentId(seriesId, seasonNumber, episodeNumber) {
  return `${seriesId}_s${seasonNumber}_e${episodeNumber}`;
}

/**
 * @typedef {object} EpisodeContentId
 * @property {string} seriesId
 * @property {number} seasonNumber
 * @property {number} episodeNumber
 */

/**
 * Parses an episode playback identifier without treating it as a catalog ObjectId.
 *
 * @param {string} contentId
 * @returns {EpisodeContentId | null}
 */
export function parseEpisodeContentId(contentId) {
  const match = EPISODE_CONTENT_ID_PATTERN.exec(contentId);
  if (!match?.groups) return null;

  const seasonNumber = Number(match.groups.seasonNumber);
  const episodeNumber = Number(match.groups.episodeNumber);

  if (
    !Number.isSafeInteger(seasonNumber) ||
    !Number.isSafeInteger(episodeNumber) ||
    seasonNumber < 1 ||
    episodeNumber < 1
  ) {
    return null;
  }

  return {
    seriesId: match.groups.seriesId,
    seasonNumber,
    episodeNumber,
  };
}

/**
 * Resolves which catalog resource owns a playback/progress identifier.
 *
 * @param {string} contentId
 * @returns {
 *   | { contentId: string, contentType: 'series', episode: EpisodeContentId }
 *   | { contentId: string, contentType: 'movie', episode: null }
 * }
 */
export function resolveCatalogReference(contentId) {
  const episode = parseEpisodeContentId(contentId);

  if (episode) {
    return {
      contentId: episode.seriesId,
      contentType: 'series',
      episode,
    };
  }

  return {
    contentId,
    contentType: 'movie',
    episode: null,
  };
}

/**
 * @param {string} contentId
 * @param {string} title
 */
export function buildEpisodeWatchHref(contentId, title) {
  const params = new URLSearchParams({ type: 'series' });
  if (title) params.set('title', title);

  return `/watch/${encodeURIComponent(contentId)}?${params.toString()}`;
}
