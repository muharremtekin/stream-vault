import BrowseSections from '@/components/browse/browse-sections';
import HeroBannerSection from '@/components/browse/hero-banner';

export default function BrowsePage() {
  return (
    <main className="min-h-screen bg-background">
      <HeroBannerSection />
      <div className="relative z-10 -mt-16 pb-16">
        <BrowseSections />
      </div>
    </main>
  );
}
