import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';

import { ProfileGrid } from '@/components/auth/profile-grid';

export default async function ProfileSelectPage() {
  const cookieStore = await cookies();
  const token = cookieStore.get('sv-access-token');
  if (!token?.value) {
    redirect('/login');
  }

  return (
    <div className="w-full max-w-3xl px-4">
      <ProfileGrid />
    </div>
  );
}
