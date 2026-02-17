import { GenreContent } from '@/components/browse/genre-content';

export default async function GenrePage(props: { params: Promise<{ slug: string }> }) {
  const { slug } = await props.params;

  return <GenreContent slug={slug} />;
}
