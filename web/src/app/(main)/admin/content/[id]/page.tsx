import { getTranslations } from 'next-intl/server';

import { ContentDetail } from '@/components/admin/content-detail';

export async function generateMetadata() {
  const t = await getTranslations('admin.contentDetail');
  return { title: t('pageTitle') };
}

export default async function AdminContentDetailPage(
  props: { params: Promise<{ id: string }> }
) {
  const { id } = await props.params;

  return <ContentDetail contentId={id} />;
}
