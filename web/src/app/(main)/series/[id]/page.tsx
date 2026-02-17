import ContentHero from '@/components/content/content-hero';
import ContentInfo from '@/components/content/content-info';
import SeriesEpisodesSection from '@/components/content/series-episodes-section';
import SimilarContent from '@/components/content/similar-content';

export default async function SeriesPage(props: { params: Promise<{ id: string }> }) {
  const { id } = await props.params;

  return (
    <>
      <ContentHero id={id} contentType="series" />
      <div className="px-4 sm:px-8 lg:px-12 py-10 space-y-10">
        <ContentInfo id={id} contentType="series" />
        <SeriesEpisodesSection seriesId={id} />
        <SimilarContent contentId={id} />
      </div>
    </>
  );
}
