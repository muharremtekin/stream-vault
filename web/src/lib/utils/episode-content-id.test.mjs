import assert from 'node:assert/strict';
import test from 'node:test';

import {
  buildEpisodeContentId,
  buildEpisodeWatchHref,
  parseEpisodeContentId,
  resolveCatalogReference,
} from './episode-content-id.js';

const SERIES_ID = '507f1f77bcf86cd799439011';

test('episode identifiers round-trip and resolve to their parent series', () => {
  const contentId = buildEpisodeContentId(SERIES_ID, 2, 7);

  assert.equal(contentId, `${SERIES_ID}_s2_e7`);
  assert.deepEqual(parseEpisodeContentId(contentId), {
    seriesId: SERIES_ID,
    seasonNumber: 2,
    episodeNumber: 7,
  });
  assert.deepEqual(resolveCatalogReference(contentId), {
    contentId: SERIES_ID,
    contentType: 'series',
    episode: {
      seriesId: SERIES_ID,
      seasonNumber: 2,
      episodeNumber: 7,
    },
  });
});

test('movie identifiers continue to resolve as movies', () => {
  assert.deepEqual(resolveCatalogReference(SERIES_ID), {
    contentId: SERIES_ID,
    contentType: 'movie',
    episode: null,
  });
});

test('malformed or zero-valued episode identifiers are rejected', () => {
  assert.equal(parseEpisodeContentId(`${SERIES_ID}_s0_e1`), null);
  assert.equal(parseEpisodeContentId(`${SERIES_ID}_s1_e0`), null);
  assert.equal(parseEpisodeContentId(`not-an-object-id_s1_e1`), null);
  assert.equal(parseEpisodeContentId(`${SERIES_ID}_s1`), null);
});

test('episode watch links retain the playback id and series type', () => {
  assert.equal(
    buildEpisodeWatchHref(`${SERIES_ID}_s2_e7`, 'Finale & More'),
    `/watch/${SERIES_ID}_s2_e7?type=series&title=Finale+%26+More`,
  );
});
