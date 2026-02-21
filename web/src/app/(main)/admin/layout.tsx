import { Sidebar } from '@/components/layout/sidebar';
import { AdminGuard } from '@/components/admin/admin-guard';

export default function AdminLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <AdminGuard>
      <div className="flex">
        <Sidebar />
        <div className="min-h-[calc(100vh-4rem)] w-full lg:pl-56">
          <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
            {children}
          </div>
        </div>
      </div>
    </AdminGuard>
  );
}
