# StreamVault — Faz 6: Eksik Sayfalar, İyileştirmeler & Hata Düzeltmeleri

> **Durum:** Faz 5 tamamlandı (Next.js 16 web arayüzü). Faz 6 ile eksik sayfalar oluşturulacak, mevcut hata ve uyumsuzluklar giderilecek, UX iyileştirmeleri yapılacak.

---

## 1. Eksik Sayfa Implementasyonları

### Görev 1: `/movies` — Film Listeleme Sayfası
- [ ] `web/src/app/(main)/movies/page.tsx` oluştur
- [ ] Mevcut `getMovies(params)` API fonksiyonunu kullan (sayfalama + filtreleme destekli)
- [ ] Tür filtreleme (genre tags), yıl filtreleme, sıralama seçenekleri ekle
- [ ] Mevcut `ContentCard` bileşenini kullanarak responsive grid (2-6 kolon) oluştur
- [ ] Sayfalama veya infinite scroll implementasyonu
- [ ] Loading skeleton'ları ve hata durumu (error boundary)
- [ ] `pages.movies.title` ve `pages.movies.noContent` çevirilerini kullan
- [ ] SSR/SSG uyumlu metadata (`generateMetadata`)

### Görev 2: `/series` — Dizi Listeleme Sayfası
- [ ] `web/src/app/(main)/series/page.tsx` oluştur (mevcut `series/[id]/page.tsx` ile çakışmayacak şekilde)
- [ ] Mevcut `getSeries(params)` API fonksiyonunu kullan
- [ ] Film sayfasıyla aynı filtreleme/sıralama/sayfalama altyapısı
- [ ] Responsive grid + `ContentCard` bileşeni
- [ ] Loading skeleton'ları ve hata durumu
- [ ] `pages.series.title` ve `pages.series.noContent` çevirilerini kullan
- [ ] SSR/SSG uyumlu metadata

### Görev 3: `/about` — Hakkında Sayfası
- [ ] `web/src/app/(main)/about/page.tsx` oluştur
- [ ] StreamVault tanıtım metni, özellikler, ekip bilgisi (statik içerik)
- [ ] `pages.about.title` ve `pages.about.description` çevirilerini kullan
- [ ] Responsive tasarım, Netflix-dark temaya uygun

### Görev 4: `/help` — Yardım Merkezi Sayfası
- [ ] `web/src/app/(main)/help/page.tsx` oluştur
- [ ] SSS (accordion) bileşeni, kategori bazlı yardım konuları
- [ ] İletişim bilgileri bölümü
- [ ] `pages.help.title`, `pages.help.description`, `pages.help.contact` çevirilerini kullan

### Görev 5: `/terms` — Kullanım Koşulları Sayfası
- [ ] `web/src/app/(main)/terms/page.tsx` oluştur
- [ ] Statik içerik, bölüm başlıkları ile yapılandırılmış
- [ ] `pages.terms.title` ve `pages.terms.description` çevirilerini kullan

### Görev 6: `/privacy` — Gizlilik Politikası Sayfası
- [ ] `web/src/app/(main)/privacy/page.tsx` oluştur
- [ ] Statik içerik, bölüm başlıkları ile yapılandırılmış
- [ ] `pages.privacy.title` ve `pages.privacy.description` çevirilerini kullan

---

## 2. Ortak Film/Dizi Listeleme Altyapısı

### Görev 7: Paylaşılan Catalog Grid Bileşeni
- [ ] `web/src/components/catalog/catalog-grid.tsx` oluştur — film ve dizi sayfaları ortak kullanacak
- [ ] Props: `fetchFn`, `contentType`, `emptyMessage`
- [ ] Tür filtreleme (mevcut `GenreTags` veya yeni filter bar)
- [ ] Sıralama dropdown (yeni eklenenler, en çok beğenilenler, yıl, A-Z)
- [ ] Sayfalama kontrolleri (önceki/sonraki veya infinite scroll)
- [ ] URL query parametreleriyle senkron filtre state'i (`?genre=action&sort=rating&page=2`)

### Görev 8: Catalog Hooks
- [ ] `web/src/lib/hooks/use-catalog.ts` oluştur
- [ ] `useMovieCatalog(params)` — `getMovies` etrafında TanStack Query wrapper
- [ ] `useSeriesCatalog(params)` — `getSeries` etrafında TanStack Query wrapper
- [ ] Filtre/sayfa değişiminde otomatik refetch
- [ ] `keepPreviousData` ile sayfa geçişlerinde flicker önleme

---

## 3. Backend-Frontend Uyumsuzluk Düzeltmeleri

### Görev 9: CORS Yapılandırma Düzeltmeleri ✅
- [x] Gateway CORS middleware'ine `X-Profile-Id` header'ını ekle
- [x] Backend servislerden (user-service, catalog-service, subscription-service) duplike CORS middleware'i kaldır
- [x] Gateway reverse proxy'sine upstream CORS header temizleme (`stripAndLog`) ekle
- [x] Tek kaynak CORS: yalnızca gateway CORS header'ı yönetir

### Görev 10: ProfileIcon Enum Uyumsuzluğu ✅
- [x] Backend `ProfileIcon` enum'ını frontend icon isimleriyle eşle (Avatar1→Smile, Avatar2→Cat, vb.)
- [x] `JsonStringEnumConverter` ile case-insensitive JSON deserialization
- [x] Handler'lardaki `.ToString()` çıktısını `.ToLowerInvariant()` ile lowercase yap
- [x] Frontend ve backend arasında `smile`, `cat`, `dog`, `bird`, `fish`, `rabbit`, `star`, `heart`, `ghost`, `rocket` standardı

---

## 4. UX İyileştirmeleri

### Görev 11: "Yeni Eklenenler" Özel Bölümü
- [ ] Browse sayfasına veya ayrı sayfaya "Yeni Eklenenler" bölümü ekle
- [ ] Son 30 gün içinde eklenen içerikleri listele
- [ ] `sort=createdAt` parametresiyle katalog API'den çek

### Görev 12: "En Çok Beğenilenler" Özel Bölümü
- [ ] Browse sayfasına veya ayrı sayfaya "En Çok Beğenilenler" bölümü ekle
- [ ] `sort=rating` parametresiyle katalog API'den çek
- [ ] Rating badge'i ile ContentCard zenginleştirmesi

### Görev 13: Gelişmiş Arama Filtreleri
- [ ] Yıl aralığı için esnek input (sadece preset yerine custom range)
- [ ] Çoklu tür seçimi desteği
- [ ] Filtre durumunu URL query params ile persist et
- [ ] Filtre sıfırlama butonu iyileştirmesi

### Görev 14: Watchlist İyileştirmeleri
- [ ] Sıralama seçenekleri (eklenme tarihi, A-Z, yıl)
- [ ] İçerik türüne göre filtreleme (film/dizi)
- [ ] Boş watchlist durumunda yönlendirme CTA iyileştirmesi

---

## 5. Performans & Erişilebilirlik

### Görev 15: Image Optimization
- [ ] Tüm thumbnail ve banner resimleri için Next.js `Image` bileşeni ile `sizes` prop doğrulaması
- [ ] Placeholder blur/skeleton kullanımını tüm resimlere yay
- [ ] Lazy loading ayarlarını doğrula (viewport dışı resimler)

### Görev 16: Error Boundary İyileştirmeleri
- [ ] Tüm sayfa gruplarına (`movies`, `series`, `catalog`) error.tsx ekle
- [ ] Retry mekanizması ile kullanıcı dostu hata sayfaları
- [ ] API timeout durumlarında anlamlı mesajlar

### Görev 17: Erişilebilirlik (a11y) Denetimi
- [ ] Tüm interaktif öğelerde `aria-label` doğrulaması
- [ ] Klavye navigasyonu testi (Tab, Enter, Escape)
- [ ] Ekran okuyucu uyumluluğu (NVDA/VoiceOver)
- [ ] Renk kontrastı kontrolü (WCAG AA)

---

## 6. Web Container Rebuild & Deploy

### Görev 18: Web Container Güncelleme
- [ ] Tüm sayfa değişikliklerinden sonra web container'ını rebuild et
- [ ] `docker compose build web && docker compose up -d web`
- [ ] Tüm yeni sayfaların production build'de çalıştığını doğrula
- [ ] i18n çevirilerinin her iki dilde (tr/en) doğru render edildiğini kontrol et

---

## Teknik Notlar

### Mevcut Mimari (Faz 5'ten)
- **Framework:** Next.js 16 (App Router, Turbopack)
- **State:** Zustand + TanStack Query v5
- **Styling:** Tailwind CSS 4 (Netflix-dark tema)
- **i18n:** next-intl (tr/en)
- **API:** Axios client, interceptors ile JWT token/refresh, X-Profile-Id header
- **Video:** hls.js adaptive bitrate

### Mevcut Bileşen Altyapısı
| Bileşen | Konum | Açıklama |
|---------|-------|----------|
| `ContentCard` | `components/browse/content-card.tsx` | Poster kartı, hover efekt, progress bar |
| `ContentRow` | `components/browse/content-row.tsx` | Yatay kaydırmalı içerik satırı |
| `GenreContent` | `components/browse/genre-content.tsx` | Tür bazlı grid + infinite scroll |
| `WatchlistGrid` | `components/browse/watchlist-grid.tsx` | Watchlist grid layout |
| `HeroBanner` | `components/browse/hero-banner.tsx` | Öne çıkan içerik banner |

### Mevcut API Fonksiyonları
| Fonksiyon | Dosya | Açıklama |
|-----------|-------|----------|
| `getMovies(params)` | `lib/api/catalog.ts` | Film listesi (sayfalama + filtre) |
| `getSeries(params)` | `lib/api/catalog.ts` | Dizi listesi (sayfalama + filtre) |
| `getGenres()` | `lib/api/catalog.ts` | Tür listesi |
| `getGenreContent(slug)` | `lib/api/catalog.ts` | Türe göre içerik |
| `search(params)` | `lib/api/search.ts` | Tam metin arama + facet |
| `getTrending()` | `lib/api/search.ts` | Trend içerikler |

### CatalogParams Interface
```typescript
interface CatalogParams {
  page?: number;
  pageSize?: number;
  genre?: string;
  sort?: string;
  year?: number;
}
```

### Çeviri Anahtarları (Eklendi)
`messages/tr.json` ve `messages/en.json` dosyalarına `pages` bölümü eklendi:
- `pages.movies.title` / `pages.movies.noContent`
- `pages.series.title` / `pages.series.noContent`
- `pages.about.title` / `pages.about.description`
- `pages.help.title` / `pages.help.description` / `pages.help.contact`
- `pages.terms.title` / `pages.terms.description`
- `pages.privacy.title` / `pages.privacy.description`
