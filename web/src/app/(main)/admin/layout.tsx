import { AdminGuard } from '@/components/admin/admin-guard';
import { AdminLayoutShell } from '@/components/layout/admin-layout-shell';

export default function AdminLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <AdminGuard>
      <AdminLayoutShell>{children}</AdminLayoutShell>
    </AdminGuard>
  );
}
