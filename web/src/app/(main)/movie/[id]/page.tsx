import ContentHero from '@/components/content/content-hero';
import ContentInfo from '@/components/content/content-info';
import SimilarContent from '@/components/content/similar-content';

export default async function MoviePage(props: { params: Promise<{ id: string }> }) {
  const { id } = await props.params;

  return (
    <>
      <ContentHero id={id} contentType="movie" />
      <div className="px-4 sm:px-8 lg:px-12 py-10 space-y-10">
        <ContentInfo id={id} contentType="movie" />
        <SimilarContent contentId={id} />
      </div>
    </>
  );
}
