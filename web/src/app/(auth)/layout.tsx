import Link from 'next/link';

export default function AuthLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="relative min-h-screen bg-gradient-to-br from-background via-card to-background">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_top,rgba(229,9,20,0.15),transparent_50%)]" />
      <div className="relative z-10 flex min-h-screen flex-col">
        <header className="flex justify-center pt-8">
          <Link
            href="/"
            className="text-4xl font-bold tracking-tight text-primary"
          >
            StreamVault
          </Link>
        </header>
        <main className="flex flex-1 items-center justify-center px-4 py-8">
          {children}
        </main>
      </div>
    </div>
  );
}
