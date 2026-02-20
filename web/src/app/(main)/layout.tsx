import { Navbar } from '@/components/layout/navbar';
import { Footer } from '@/components/layout/footer';
import { NotificationProvider } from '@/components/notification/notification-provider';

export default function MainLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <>
      <Navbar />
      <NotificationProvider>
        <main className="min-h-screen pt-16">{children}</main>
      </NotificationProvider>
      <Footer />
    </>
  );
}
