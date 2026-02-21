# StreamVault — Faz 5: Frontend Web Uygulaması Görevleri

> **Durum:** Faz 1-4 tamamlandı (8 backend servis + observability altyapısı). Faz 5 ile Next.js 16 tabanlı Netflix görünümlü bir web arayüzü oluşturulacak. Tüm backend endpoint'lerinin karşılığı olan UI bileşenleri implemente edilecek.

---

## Teknoloji Kararları

| Kriter | Karar | Sebep |
|--------|-------|-------|
| Framework | Next.js 16 (App Router + Turbopack) | SSR/SSG, Turbopack varsayılan, React 19.2 |
| Dil | TypeScript 5+ | Backend proto/DTO'larıyla uyumlu tip güvenliği |
| Runtime | Node.js 20.9+ | Next.js 16 minimum gereksinim |
| Styling | Tailwind CSS 4 | Hızlı geliştirme, Netflix-dark tema |
| State | Zustand | Basit, boilerplate'siz global state |
| Data Fetching | TanStack Query v5 (React Query) | Cache, retry, optimistic update |
| Video Player | hls.js | HLS adaptive bitrate playback |
| WebSocket | Native WebSocket + reconnect | Bildirimler için |
| Form | React Hook Form + Zod | Validasyonlu formlar |
| Icons | Lucide React | Hafif, tutarlı icon seti |
| Toast/Alert | Sonner | Bildirim toast'ları |
| Auth Redirect | `proxy.ts` (Next.js 16) | `middleware.ts` yerine yeni convention |
| Config | `next.config.ts` | TypeScript config (Next.js 16 standart) |
| Bundler | Turbopack (varsayılan) | Next.js 16'da default, `--turbopack` flag'i gereksiz |
| ESLint | ESLint Flat Config (`eslint.config.mjs`) | `next lint` kaldırıldı, ESLint CLI doğrudan kullanılır |
| i18n | next-intl | Server/Client component desteği, ICU format, type-safe |
| React | React 19.2 | View Transitions, useEffectEvent, Activity |

---

## 1. Proje Kurulumu & Altyapı

### 1.1 Next.js 16 Proje Oluşturma
- [x] `web/` dizini oluştur
- [x] `npx create-next-app@latest` ile Next.js 16 projesi oluştur (TypeScript, Tailwind CSS, App Router, src/ dizini)
- [x] Node.js 20.9+ gereksinimini doğrula
- [x] `next.config.ts` oluştur (TypeScript config)
- [x] Turbopack'in varsayılan olarak çalıştığını doğrula (`next dev` → Turbopack aktif)
- [x] `package.json` script'leri: `dev: "next dev"`, `build: "next build"`, `start: "next start"` (flag'siz)
- [x] `.env.local.example` oluştur (`NEXT_PUBLIC_API_URL`, `NEXT_PUBLIC_WS_URL`)

### 1.2 Bağımlılıklar
- [x] Zustand ekle (`zustand`)
- [x] TanStack Query v5 ekle (`@tanstack/react-query`, `@tanstack/react-query-devtools`)
- [x] Axios ekle (`axios`)
- [x] React Hook Form + Zod ekle (`react-hook-form`, `@hookform/resolvers`, `zod`)
- [x] hls.js ekle (`hls.js`)
- [x] Lucide React ekle (`lucide-react`)
- [x] Sonner ekle (`sonner`)
- [x] clsx + tailwind-merge ekle (`clsx`, `tailwind-merge`)
- [x] next-intl ekle (`next-intl`)

### 1.3 ESLint & Kod Kalitesi
- [x] `eslint.config.mjs` oluştur (Flat Config formatı — `next lint` kaldırıldı)
- [x] `@next/eslint-plugin-next` ekle
- [x] TypeScript ESLint kuralları ekle
- [x] Prettier entegrasyonu (opsiyonel)
- [x] `.prettierrc` oluştur

### 1.4 Docker Entegrasyonu
- [x] `web/Dockerfile` oluştur (multi-stage: node:20-alpine builder + runner)
- [x] Build stage: `npm ci` → `npm run build`
- [x] Runner stage: `.next`, `public`, `package.json`, `node_modules` kopyala
- [x] `EXPOSE 3001`, `CMD ["npm", "start"]`
- [x] `docker-compose.yml`'a `web` servisi ekle (port: 3001:3001)
- [x] Environment: `NEXT_PUBLIC_API_URL=http://localhost:8081`, `NEXT_PUBLIC_WS_URL=ws://localhost:8081`
- [x] `depends_on: [gateway]`
- [ ] `docker compose up web` ile frontend'in ayağa kalktığını doğrula

### 1.5 Utility Fonksiyonlar
- [x] `src/lib/utils/cn.ts` — Tailwind class merge (`clsx` + `tailwind-merge`)
- [x] `src/lib/utils/format.ts` — Süre formatı (120dk → "2s 0dk"), tarih formatı, para formatı
- [x] `src/lib/utils/constants.ts` — API URL, WS URL, renk kodları, tier limitleri

### 1.6 i18n Kurulumu (next-intl)
- [x] `src/i18n/request.ts` oluştur — `getRequestConfig` ile locale ve mesaj dosyası yükleme
- [x] `messages/en.json` oluştur — İngilizce çeviriler (default locale)
- [x] `messages/tr.json` oluştur — Türkçe çeviriler
- [x] `src/app/layout.tsx`'e `NextIntlClientProvider` ekle (mesajları server'dan client'a aktar)
- [x] `next.config.ts`'e `createNextIntlPlugin` entegrasyonu
- [x] Başlangıç namespace'leri: `common` (genel butonlar, hatalar), `auth` (login/register), `browse` (ana sayfa)
- [x] Locale algılama: `Accept-Language` header veya kullanıcı tercihi (cookie/localStorage)
- [x] Tüm UI metinleri bu noktadan itibaren `useTranslations()` ile yazılır — hardcoded metin YASAKTIR

---

## 2. Tema & Layout Sistemi

### 2.1 Dark Tema & Global Stiller
- [x] `src/app/globals.css` — Tailwind import'ları, dark tema renk paleti (Netflix-dark)
- [x] CSS custom properties: `--background`, `--foreground`, `--primary`, `--accent`, `--muted` vb.
- [x] Font ayarı (Inter veya benzeri sans-serif)
- [x] Scrollbar stili (dark tema uyumlu)
- [x] Tailwind config'de custom renkleri tanımla

### 2.2 Root Layout
- [x] `src/app/layout.tsx` — HTML lang, dark tema class, font, metadata
- [x] `<Toaster />` (Sonner) provider ekle
- [x] TanStack Query `QueryClientProvider` ekle
- [x] Global error boundary

### 2.3 Auth Layout Grubu
- [x] `src/app/(auth)/layout.tsx` — Navbar yok, arka plan bulanık poster collage
- [x] Logo üstte ortalanmış
- [x] Sadece login/register sayfaları bu layout'u kullanır

### 2.4 Main Layout Grubu
- [x] `src/app/(main)/layout.tsx` — Navbar + NotificationProvider + Footer
- [x] Tüm korumalı sayfalar bu layout'u kullanır

### 2.5 Layout Bileşenleri
- [x] `src/components/layout/navbar.tsx` — Logo, navigasyon linkleri (Ana Sayfa, Diziler, Filmler, Listem), arama ikonu, bildirim bell, profil dropdown
- [x] `src/components/layout/footer.tsx` — Basit footer
- [x] `src/components/layout/sidebar.tsx` — Admin sidebar (Dashboard, İçerikler, Encoding)
- [x] `src/components/layout/profile-switcher.tsx` — Profil değiştirme dropdown (navbar içinde)
- [x] `src/components/layout/notification-bell.tsx` — Bildirim ikonu + unread count badge + dropdown

---

## 3. Temel UI Bileşenleri

- [x] `src/components/ui/button.tsx` — Primary, secondary, ghost, danger varyantları
- [x] `src/components/ui/input.tsx` — Text input, label, error mesajı desteği
- [x] `src/components/ui/modal.tsx` — Overlay modal (ESC ile kapanır)
- [x] `src/components/ui/dropdown.tsx` — Dropdown menü
- [x] `src/components/ui/skeleton.tsx` — Loading skeleton (card, text, image varyantları)
- [x] `src/components/ui/badge.tsx` — Renkli badge (tür, durum vb.)
- [x] `src/components/ui/tooltip.tsx` — Hover tooltip
- [x] `src/components/ui/progress-bar.tsx` — İlerleme çubuğu (izleme, upload, encoding)
- [x] `src/components/ui/toast.tsx` — Sonner toast wrapper (gerekirse)

---

## 4. API Client & State Management

### 4.1 TypeScript Tipleri
- [x] `src/lib/types/common.ts` — `PaginatedResponse<T>`, `ApiError`, `ApiResponse<T>`
- [x] `src/lib/types/auth.ts` — `User`, `Profile`, `LoginRequest`, `LoginResponse`, `RegisterRequest`
- [x] `src/lib/types/catalog.ts` — `Movie`, `Series`, `Episode`, `Season`, `Genre`, `VideoStatus`
- [x] `src/lib/types/streaming.ts` — `StreamingInfo`, `WatchProgress`, `ContinueWatchingItem`
- [x] `src/lib/types/search.ts` — `SearchHit`, `SearchResponse`, `Facets`, `FacetBucket`, `AutocompleteItem`
- [x] `src/lib/types/recommendation.ts` — `RecommendedItem`, `SimilarItem`, `HomeSection`
- [x] `src/lib/types/subscription.ts` — `Plan`, `Subscription`, `Invoice`, `PaymentRequest`
- [x] `src/lib/types/notification.ts` — `Notification`, `NotificationPreferences`

### 4.2 API Client
- [x] `src/lib/api/client.ts` — Axios instance oluştur (`baseURL: NEXT_PUBLIC_API_URL`, timeout: 10s)
- [x] Request interceptor: `Authorization: Bearer {token}` header ekle (Zustand store'dan)
- [x] Request interceptor: `X-Profile-Id` header ekle (aktif profil)
- [x] Response interceptor: 401 → refresh token dene → başarısızsa logout + `/login`'e redirect
- [x] `src/lib/api/auth.ts` — `login()`, `register()`, `refreshToken()`, `getProfiles()`, `createProfile()`
- [x] `src/lib/api/catalog.ts` — `getMovies()`, `getMovie(id)`, `getSeries()`, `getSeriesById(id)`, `getGenres()`, `getGenreContent(slug)`
- [x] `src/lib/api/streaming.ts` — `getStreamingInfo(id)`, `getProgress(id)`, `saveProgress(id, position)`, `getContinueWatching()`, `uploadVideo(formData, onProgress)`
- [x] `src/lib/api/search.ts` — `search(query, filters)`, `autocomplete(query)`, `getTrending(window)`
- [x] `src/lib/api/recommendation.ts` — `getHomeSections()`, `getSimilar(id)`, `getRecommendations()`
- [x] `src/lib/api/subscription.ts` — `getPlans()`, `getMySubscription()`, `subscribe(planId, payment)`, `changePlan(planId)`, `cancel()`, `getInvoices()`
- [x] `src/lib/api/notification.ts` — `getNotifications(params)`, `markAsRead(id)`, `markAllAsRead()`, `getPreferences()`, `updatePreferences(prefs)`
- [x] `src/lib/api/encoding.ts` — `getJobs(params)`, `getJob(id)` (admin)

### 4.3 Zustand Store'lar
- [x] `src/lib/stores/auth-store.ts` — `user`, `accessToken`, `refreshToken`, `activeProfile`, `profiles`, `subscriptionTier`
- [x] Auth store actions: `login()`, `register()`, `logout()`, `refreshAuth()`, `setActiveProfile()`, `fetchProfiles()`
- [x] `persist` middleware ile token'ları localStorage'a kaydet
- [x] `src/lib/stores/player-store.ts` — `contentId`, `isPlaying`, `currentTime`, `duration`, `currentQuality`, `availableQualities`, `volume`, `isFullscreen`, `isBuffering`
- [x] Player store actions: `setQuality()`, `togglePlay()`, `seek()`, `setVolume()`
- [x] `src/lib/stores/notification-store.ts` — `notifications`, `unreadCount`, `wsConnected`
- [x] Notification store actions: `connect()`, `disconnect()`, `markAsRead()`, `markAllAsRead()`, `fetchHistory()`

### 4.4 Custom Hooks (TanStack Query)
- [x] `src/lib/hooks/use-auth.ts` — Auth state ve işlemleri (store wrapper)
- [x] `src/lib/hooks/use-profile.ts` — Aktif profil yönetimi
- [x] `src/lib/hooks/use-search.ts` — Debounced arama (300ms) + autocomplete query
- [x] `src/lib/hooks/use-player.ts` — Player state + progress save (15s interval)
- [x] `src/lib/hooks/use-watchlist.ts` — Watchlist CRUD (TanStack Query mutations)
- [x] `src/lib/hooks/use-notifications.ts` — WebSocket bağlantı + bildirim state
- [x] `src/lib/hooks/use-subscription.ts` — Abonelik durumu query

---

## 5. Auth & Profil Sayfaları

### 5.1 Landing Page (`/`)
- [x] `src/app/page.tsx` — Statik landing page
- [x] Logo + "Giriş Yap" butonu üstte
- [x] Hero section: başlık, açıklama, email input + "Başla" butonu
- [x] Özellik kartları (3 kart: TV, mobil, profil)
- [x] SSS accordion bölümü
- [x] Giriş yapmışsa → `/browse`'a redirect (server-side kontrol)

### 5.2 Login Sayfası (`/login`)
- [x] `src/app/(auth)/login/page.tsx`
- [x] `src/components/auth/login-form.tsx` — React Hook Form + Zod validasyon
- [x] Email + şifre input'ları
- [x] "Giriş Yap" submit butonu
- [x] "Hesabın yok mu? Kayıt ol" linki
- [x] Loading state (submit sırasında)
- [x] Hata gösterimi (yanlış şifre, kullanıcı bulunamadı)
- [x] Başarılı login → profil seçimine veya `/browse`'a yönlendir
- [x] API: `POST /api/auth/login`

### 5.3 Register Sayfası (`/register`)
- [x] `src/app/(auth)/register/page.tsx`
- [x] `src/components/auth/register-form.tsx` — React Hook Form + Zod validasyon
- [x] Email, şifre, şifre tekrarı input'ları
- [x] Şifre gücü göstergesi (opsiyonel)
- [x] Başarılı kayıt → login sayfasına yönlendir veya otomatik giriş
- [x] API: `POST /api/auth/register`

### 5.4 Profil Seçimi
- [x] `src/components/auth/profile-select.tsx` — "Kim izliyor?" ekranı
- [x] Profil kartları grid'i (avatar + isim)
- [x] "Profil Ekle" kartı (+ ikonu)
- [x] Yeni profil oluşturma modal'ı (isim + avatar seçimi)
- [x] Max 5 profil sınırı UI'da gösterilmeli
- [x] Profil seçildiğinde → `setActiveProfile()` + `/browse`'a yönlendir
- [x] "Profilleri Yönet" butonu
- [x] API: `GET /api/users/me/profiles`, `POST /api/users/me/profiles`

### 5.5 Auth Proxy (Korumalı Route'lar)
- [x] `src/proxy.ts` oluştur (Next.js 16 — `middleware.ts` yerine `proxy.ts`)
- [x] Named export: `export function proxy(request: Request)`
- [x] Token yoksa korumalı route'ları `/login`'e redirect et
- [x] Public route'lar: `/`, `/login`, `/register`
- [x] Matcher config: `/((?!_next/static|_next/image|favicon.ico).*)` patterni

---

## 6. Browse & İçerik Sayfaları

### 6.1 Ana Sayfa (`/browse`)
- [x] `src/app/(main)/browse/page.tsx`
- [x] `src/components/browse/hero-banner.tsx` — Rastgele öne çıkan içerik (büyük banner)
- [x] Banner: poster arka plan, gradient overlay, başlık, rating, yıl, süre, açıklama
- [x] Banner butonları: "Oynat", "Listem", "Detay"
- [x] `src/components/browse/content-row.tsx` — Yatay kaydırmalı içerik sırası (ok butonları ile scroll)
- [x] Section başlığı (ör: "Sizin İçin Seçtiklerimiz", "Türkiye'de Trend")
- [x] Sağa/sola kaydırma ok butonları
- [x] "Kaldığın Yerden Devam Et" sırası — progress bar'lı kartlar
- [x] "Sizin İçin Seçtiklerimiz" sırası
- [x] "Trend" sırası (1-10 numaralı)
- [x] İzleme geçmişine dayalı öneri sıraları
- [ ] Tür bazlı sıralar (Bilim Kurgu, Dram, Aksiyon vb.)
- [x] "Yeni Eklenenler" sırası
- [x] API: `GET /api/recommendations/home`, `GET /api/stream/continue-watching`, `GET /api/search/trending?window=week`

### 6.2 Content Card
- [x] `src/components/browse/content-card.tsx` — Poster + başlık kartı
- [x] Poster image (fallback placeholder)
- [x] Başlık alt yazısı
- [x] `src/components/browse/content-card-hover.tsx` — Hover detay kartı
- [x] 300ms delay ile hover aktifleşir
- [x] Kart `scale(1.3)` ile büyür, `z-index` ile üste çıkar
- [x] Poster (zoom), başlık, rating, yıl, türler, süre
- [x] Aksiyon butonları: Oynat, Listeye Ekle, Beğen, Detay
- [x] Viewport kenarında sola/sağa kayma (taşma önleme)
- [x] `src/components/browse/genre-tags.tsx` — Tür etiketleri
- [x] `src/components/browse/trending-badge.tsx` — "#1 Trend" rozeti

### 6.3 Film Detay (`/movie/[id]`)
- [x] `src/app/(main)/movie/[id]/page.tsx` — Async params: `const { id } = await props.params`
- [x] `src/components/content/content-hero.tsx` — Banner arka plan + gradient + bilgi
- [x] Başlık, rating (oy sayısı), yıl, yaş sınıfı, süre
- [x] "Oynat" ve "Listem" butonları
- [x] `src/components/content/content-info.tsx` — Açıklama, yönetmen, oyuncular, türler
- [x] `src/components/content/rating-stars.tsx` — 5 yıldız puanlama (görsel)
- [x] `src/components/content/maturity-badge.tsx` — Yaş sınıfı rozeti (PG-13, R vb.)
- [x] `src/components/content/add-to-list-button.tsx` — Watchlist toggle butonu (ekle/çıkar)
- [x] `src/components/content/similar-content.tsx` — Benzerleri satırı
- [x] API: `GET /api/catalog/movies/{id}`, `GET /api/recommendations/similar/{id}`, `POST /api/users/me/profiles/{pid}/watchlist`

### 6.4 Dizi Detay (`/series/[id]`)
- [x] `src/app/(main)/series/[id]/page.tsx` — Async params
- [x] Film detayla aynı hero + info bileşenleri
- [x] `src/components/content/season-selector.tsx` — Sezon seçici (tab)
- [x] `src/components/content/episode-list.tsx` — Bölüm listesi (numara, başlık, süre, açıklama, oynat butonu)
- [x] `src/components/content/series-episodes-section.tsx` — Sezon/bölüm state yönetimi
- [ ] Her bölümde izleme progress'i göster (izlenmişse)
- [x] API: `GET /api/catalog/series/{id}`, `GET /api/recommendations/similar/{id}`

### 6.5 Watchlist (`/my-list`)
- [x] `src/app/(main)/my-list/page.tsx`
- [x] Grid görünümünde content card'lar
- [x] "Listeden Çıkar" hover aksiyonu
- [x] Liste boşsa empty state ("Henüz bir şey eklemediniz")
- [x] API: `GET /api/users/me/profiles/{pid}/watchlist`, `DELETE /api/users/me/profiles/{pid}/watchlist/{id}`

### 6.6 Genre Sayfası (`/genre/[slug]`)
- [x] `src/app/(main)/genre/[slug]/page.tsx` — Async params
- [x] Tür başlığı + content card grid
- [x] Sayfalama (load more veya infinite scroll)
- [x] API: `GET /api/catalog/genres/{slug}/content`

---

## 7. Video Player

### 7.1 Ana Player Bileşeni
- [x] `src/components/player/video-player.tsx` — hls.js wrapper
- [x] hls.js instance oluşturma ve video element'e attach
- [x] HLS manifest yükleme (`GET /stream/{id}/manifest.m3u8`)
- [x] Adaptive bitrate otomatik kalite seçimi
- [x] Error handling (network, media, fatal/non-fatal)
- [x] Cleanup: unmount'ta `hls.destroy()`
- [x] Native fallback: Safari'de HLS natively desteklendiğinde `<video>` src doğrudan kullan

### 7.2 Player Kontrolleri
- [x] `src/components/player/player-controls.tsx`
- [x] Play/Pause butonu
- [x] Seek bar (progress slider)
- [x] Geçerli zaman / toplam süre gösterimi
- [x] Volume slider + mute toggle
- [x] Fullscreen toggle butonu
- [x] Kontroller 3 saniye hareketsizlikte gizlenir (mouse move ile gösterilir)

### 7.3 Kalite Seçici
- [x] `src/components/player/quality-selector.tsx`
- [x] "Otomatik" seçenek (ABR — `hls.currentLevel = -1`)
- [x] Manuel kalite seçenekleri (360p, 720p, 1080p, 4K)
- [x] Tier bazlı kısıtlama: Basic → max 720p, Standard → max 1080p, Premium → 4K
- [x] Yetkisi olmayan kalite seviyeleri `disabled` göster
- [x] `hls.currentLevel = index` ile manuel seçim

### 7.4 İzleme Pozisyonu Takibi
- [x] `src/components/player/progress-tracker.tsx`
- [x] Sayfa açılırken: `GET /api/stream/{id}/progress` → kaldığı yere seek
- [x] Oynatma sırasında: her 15 saniyede `POST /api/stream/{id}/progress`
- [x] `beforeunload` event'inde son pozisyonu kaydet
- [x] `visibilitychange` event'inde pozisyon kaydet (tab değişimi)

### 7.5 Player Overlay & UX
- [x] `src/components/player/player-overlay.tsx` — Başlık, geri butonu
- [x] Buffering göstergesi (spinner)
- [x] Oynatma bittiğinde: dizi ise "Sonraki Bölüm" önerisi (10s countdown)
- [x] Oynatma bittiğinde: film ise "Benzerleri" önerisi
- [x] Keyboard shortcut'lar: Space (play/pause), F (fullscreen), ← (10s geri), → (10s ileri), M (mute), ↑/↓ (volume)

### 7.6 Watch Sayfası (`/watch/[contentId]`)
- [x] `src/app/(main)/watch/[contentId]/page.tsx` — Async params
- [x] Tam ekran player layout (navbar gizli)
- [x] Player bileşenlerini birleştir (video-player + controls + quality + progress + overlay)
- [x] API: `GET /stream/{id}/manifest.m3u8`, `GET /api/stream/{id}/progress`, `POST /api/stream/{id}/progress`

---

## 8. Arama

### 8.1 Search Bar
- [x] `src/components/search/search-bar.tsx` — Navbar'da arama ikonu + genişleyen input
- [x] İkon tıklandığında input açılır (animasyonlu)
- [x] 300ms debounce ile autocomplete tetiklenir
- [x] 2+ karakter sonrasında öneri listesi göster
- [x] Enter'a basınca `/search?q=...` sayfasına yönlendir

### 8.2 Autocomplete
- [x] Autocomplete dropdown (search bar altında)
- [x] Film/dizi ikonu + başlık + yıl
- [x] Tıklandığında ilgili detay sayfasına git
- [x] API: `GET /api/search/autocomplete?q=...`

### 8.3 Arama Sonuçları Sayfası (`/search`)
- [x] `src/app/(main)/search/page.tsx` — Async searchParams: `const query = (await props.searchParams).q`
- [x] `src/components/search/search-results.tsx` — Sonuç grid'i (content card'lar)
- [x] Sonuç sayısı + arama süresi gösterimi ("15 sonuç, 12ms")
- [x] `src/components/search/search-highlight.tsx` — Arama terimini vurgulu göster
- [x] Boş sonuç state'i ("Sonuç bulunamadı")
- [x] Boş arama → trending göster

### 8.4 Facet Filtreleri
- [x] `src/components/search/search-filters.tsx`
- [x] Tür filtresi (checkbox grubu + sonuç sayıları)
- [x] Yıl filtresi (aralık seçimi: 2020+, 2015-2020, 2010-2015, Daha Eski)
- [x] Rating filtresi (minimum: 7+, 8+, 9+)
- [x] Tip filtresi (Filmler, Diziler, Tümü)
- [x] Sıralama (İlgililik, Puan, Yıl, İsim)
- [x] Filtre değişikliğinde URL parametreleri güncelle
- [x] Sayfalama (sayfa numaraları veya load more)
- [x] API: `GET /api/search?q=...&genres=...&yearFrom=...&yearTo=...&minRating=...&type=...&sort=...&page=...`

---

## 9. Abonelik Yönetimi

### 9.1 Abonelik Sayfası (`/account/subscription`)
- [x] `src/app/(main)/account/subscription/page.tsx`
- [x] `src/components/subscription/subscription-status.tsx` — Aktif abonelik kartı (plan adı, fiyat, durum, yenileme tarihi, özellikler)
- [x] "Plan Değiştir" ve "İptal Et" butonları

### 9.2 Plan Kartları
- [x] `src/components/subscription/plan-cards.tsx` — 3 plan kartı (Basic, Standard, Premium)
- [x] Her kart: plan adı, aylık fiyat, özellik listesi (kalite, ekran sayısı, indirme vb.)
- [x] Mevcut plan vurgulanmış ("Mevcut Plan" badge)
- [x] Upgrade/Downgrade butonları
- [x] API: `GET /api/plans`

### 9.3 Ödeme Formu
- [x] `src/components/subscription/payment-form.tsx` — Mock kart bilgisi formu
- [x] Kart numarası, son kullanma, CVV input'ları (Zod validasyon)
- [x] "Ödemeyi Onayla" butonu
- [x] Başarılı → abonelik durumu güncelle
- [x] API: `POST /api/subscriptions`

### 9.4 Plan Değiştirme & İptal
- [x] Plan değiştirme onay modal'ı (fiyat farkı gösterimi)
- [x] İptal onay modal'ı ("Emin misiniz?")
- [x] API: `PUT /api/subscriptions/me/plan`, `POST /api/subscriptions/me/cancel`

### 9.5 Fatura Geçmişi
- [x] `src/components/subscription/invoice-table.tsx` — Fatura tablosu
- [x] Sütunlar: tarih, plan, tutar, durum
- [x] API: `GET /api/subscriptions/me/invoices`

---

## 10. Bildirimler

### 10.1 WebSocket Provider
- [x] `src/components/notification/notification-provider.tsx`
- [x] Kullanıcı login olduğunda WebSocket bağlantısı aç (`ws://.../ws/notifications?token=...`)
- [x] Gelen mesajları notification store'a ekle
- [x] Toast göster (Sonner ile — bildirim tipi bazlı ikon)
- [x] Bağlantı koparsa exponential backoff ile reconnect (1s → 2s → 4s → 8s → max 30s)
- [x] Heartbeat: her 30 saniyede ping, 10s içinde pong gelmezse reconnect
- [x] Logout'ta bağlantıyı kapat
- [x] Provider'ı `(main)/layout.tsx`'e sar

### 10.2 Bildirim Bell & Dropdown
- [x] Navbar'da bildirim ikonu
- [x] Unread count badge (kırmızı daire + sayı)
- [x] Tıklandığında dropdown açılır (son 5-10 bildirim)
- [x] "Tümünü Okundu İşaretle" butonu
- [x] "Tüm Bildirimleri Gör" linki → `/account/notifications`

### 10.3 Bildirim Listesi Sayfası
- [x] `src/app/(main)/account/notifications/page.tsx`
- [x] `src/components/notification/notification-list.tsx` — Sayfalı bildirim listesi
- [x] `src/components/notification/notification-item.tsx` — Tek bildirim kartı (ikon, mesaj, tarih, okundu/okunmadı)
- [x] Tıklandığında okundu işaretle + ilgili sayfaya yönlendir
- [x] API: `GET /api/notifications?unreadOnly=...`, `POST /api/notifications/{id}/read`, `POST /api/notifications/read-all`

### 10.4 Bildirim Tercihleri
- [x] `src/components/notification/notification-prefs.tsx` — Tercih ayarları formu
- [x] Bildirim türleri (yeni içerik, öneri, abonelik vb.) için toggle'lar
- [x] API: `GET /api/notifications/preferences`, `PUT /api/notifications/preferences`

---

## 11. Hesap Sayfaları

### 11.1 Hesap Genel Bakış (`/account`)
- [ ] `src/app/(main)/account/page.tsx`
- [ ] Kullanıcı bilgileri (email, kayıt tarihi)
- [ ] Aktif abonelik özeti
- [ ] Navigasyon: Profiller, Abonelik, Bildirimler

### 11.2 Profil Yönetimi (`/account/profiles`)
- [ ] `src/app/(main)/account/profiles/page.tsx`
- [ ] Profil listesi (düzenle/sil butonları)
- [ ] Profil düzenleme modal'ı (isim, avatar değiştirme)
- [ ] Yeni profil ekleme (max 5 sınırı)
- [ ] API: `GET /api/users/me/profiles`, `POST /api/users/me/profiles`

---

## 12. Admin Paneli

### 12.1 Admin Layout & Guard
- [x] `src/app/(main)/admin/layout.tsx` — Admin sidebar + role guard
- [x] Kullanıcının `role === "Admin"` kontrolü, değilse `/browse`'a redirect
- [x] Sidebar: Dashboard, İçerikler, Encoding linkleri

### 12.2 Admin Dashboard (`/admin`)
- [x] `src/app/(main)/admin/page.tsx`
- [x] `src/components/admin/stats-cards.tsx` — İstatistik kartları (toplam içerik, encoding işlemde, aktif abone)
- [x] Son encoding işlemleri listesi (kısa tablo: başlık, kalite, progress, durum)

### 12.3 İçerik Yönetimi (`/admin/content`)
- [x] `src/app/(main)/admin/content/page.tsx` — İçerik listesi tablosu
- [x] Sütunlar: başlık, tür (Film/Dizi), video durumu (Ready/Encoding/No Video), tarih
- [x] Arama filtresi
- [x] "Yeni Ekle" butonu → `/admin/content/new`
- [x] Satır tıklandığında → `/admin/content/[id]`
- [x] API: `GET /api/catalog/movies`, `GET /api/catalog/series`

### 12.4 İçerik Ekleme/Düzenleme (`/admin/content/new`, `/admin/content/[id]`)
- [x] `src/app/(main)/admin/content/new/page.tsx`
- [x] `src/app/(main)/admin/content/[id]/page.tsx` — Async params
- [x] `src/components/admin/content-form.tsx` — Film/dizi ekleme/düzenleme formu (React Hook Form + Zod)
- [x] Alanlar: başlık, açıklama, yıl, rating (dropdown), türler (multi-select), yönetmen, oyuncular (tag input)
- [x] Film/Dizi tipi seçimi (dizi seçilince sezon/bölüm ekleme alanları)
- [x] "Kaydet" ve "İptal" butonları
- [x] API: `POST /api/catalog/movies`, `POST /api/catalog/series`, `GET /api/catalog/movies/{id}`

### 12.5 Video Upload
- [x] `src/components/admin/video-uploader.tsx` — Drag & drop video upload bileşeni
- [x] Desteklenen formatlar gösterimi (MP4, MKV, AVI — max 10GB)
- [x] Upload progress bar (yüzde + boyut)
- [x] Upload tamamlanınca encoding job ID göster
- [x] Axios `onUploadProgress` ile gerçek zamanlı ilerleme
- [x] API: `POST /api/stream/upload` (multipart/form-data)

### 12.6 Encoding İzleme (`/admin/encoding`)
- [x] `src/app/(main)/admin/encoding/page.tsx`
- [x] `src/components/admin/encoding-status.tsx` — Encoding job listesi tablosu
- [x] Sütunlar: içerik adı, kalite, progress bar, durum (Pending/Processing/Completed/Failed)
- [x] `src/components/admin/encoding-progress.tsx` — Gerçek zamanlı ilerleme çubuğu (polling ile güncelleme, 5s interval)
- [x] Detay sayfası: job bilgileri, hata mesajı (varsa)
- [x] API: `GET /api/encoding/jobs`, `GET /api/encoding/jobs/{id}`

---

## 13. UX & Responsive

### 13.1 Loading State'ler
- [ ] Her sayfa için loading skeleton'lar (Suspense boundary veya TanStack Query `isLoading`)
- [ ] Content card skeleton (poster + text placeholder)
- [ ] Tablo skeleton (satır placeholder)
- [ ] Hero banner skeleton

### 13.2 Error State'ler
- [ ] Global error boundary (`error.tsx`)
- [ ] API hata gösterimi (toast veya inline mesaj)
- [ ] 404 sayfası (`not-found.tsx`)
- [ ] Network hata sayfası

### 13.3 Empty State'ler
- [ ] Watchlist boş ("Henüz bir şey eklemediniz")
- [ ] Arama sonuç yok ("Sonuç bulunamadı")
- [ ] Bildirim yok ("Yeni bildiriminiz yok")
- [ ] Admin içerik yok ("Henüz içerik eklenmemiş")

### 13.4 Responsive Tasarım
- [ ] Navbar: mobilde hamburger menü
- [ ] Content card grid: mobilde 2 sütun, tablette 3, masaüstünde 5-6
- [ ] Content row: mobilde tek sıra kaydırma
- [ ] Player: mobilde kontrol boyutları büyütülmüş
- [ ] Admin sidebar: mobilde overlay
- [ ] Form'lar: mobilde full-width

### 13.5 Animasyonlar & Geçişler
- [ ] React 19.2 View Transitions API ile sayfa geçiş animasyonları (opsiyonel)
- [ ] Content card hover animasyonu (scale + opacity transition)
- [ ] Modal açılış/kapanış animasyonu
- [ ] Navbar scroll'da arka plan değişimi (transparent → solid)

---

## 14. Uçtan Uca Test & Doğrulama

### 14.1 Auth Akışı
- [ ] Register → login → profil seçimi → browse akışı çalışıyor
- [ ] Token refresh otomatik çalışıyor (401 → retry)
- [ ] Logout → login'e redirect çalışıyor
- [ ] Korumalı route'lar unauthorized erişimi engelliyor

### 14.2 Browse & İçerik Akışı
- [ ] Ana sayfa tüm section'ları yüklüyor
- [ ] Content card hover efekti çalışıyor
- [ ] Film/dizi detay sayfaları tüm bilgileri gösteriyor
- [ ] Puanlama çalışıyor
- [ ] Watchlist ekleme/çıkarma çalışıyor

### 14.3 Player Akışı
- [ ] HLS video oynatılabiliyor
- [ ] Kalite seçimi çalışıyor
- [ ] İzleme pozisyonu kaydediliyor
- [ ] Kaldığın yerden devam çalışıyor
- [ ] Keyboard shortcut'lar çalışıyor

### 14.4 Arama Akışı
- [ ] Autocomplete çalışıyor (2+ karakter)
- [ ] Full-text arama sonuç dönüyor
- [ ] Facet filtreleri çalışıyor
- [ ] Boş aramada trendler gösteriliyor

### 14.5 Admin Akışı
- [ ] Admin paneline sadece admin erişebiliyor
- [ ] İçerik CRUD çalışıyor
- [ ] Video upload progress gösteriyor
- [ ] Encoding job durumu görüntüleniyor

### 14.6 Bildirim Akışı
- [ ] WebSocket bağlantısı kuruluyor
- [ ] Toast bildirimi gösteriliyor
- [ ] Bell'de unread count güncelleniyor
- [ ] Okundu işaretleme çalışıyor

### 14.7 Docker Doğrulama
- [ ] `docker compose up` ile frontend ayağa kalkıyor
- [ ] Gateway üzerinden tüm API istekleri başarılı
- [ ] HLS streaming tarayıcıdan çalışıyor

---

## Önerilen Sıralama

1. **Proje kurulumu** → Next.js 16 + TS + Tailwind + bağımlılıklar + Docker
2. **Tema & Layout** → Dark tema, root/auth/main layout'lar, navbar, footer
3. **UI bileşenleri** → Button, input, modal, skeleton, badge, progress-bar
4. **API client & state** → Axios + interceptor, Zustand store'lar, TypeScript tipleri
5. **Auth sayfaları** → Landing, login, register, profil seçimi, proxy.ts
6. **Browse — Ana sayfa** → Hero banner, content-row, content-card, recommendation API
7. **Content card hover & Detay** → Hover efekti, film detay, dizi detay (sezon/bölüm)
8. **Video player** → hls.js, kontroller, kalite seçici, progress tracking, keyboard shortcuts
9. **Arama** → Search bar, autocomplete, sonuç sayfası, facet filtreleri
10. **Watchlist & Genre** → Listem sayfası, tür sayfası
11. **Abonelik** → Plan kartları, mock ödeme, durum, değiştirme, iptal, fatura
12. **Bildirimler** → WebSocket provider, bell, dropdown, liste, tercihler, toast
13. **Admin paneli** → Layout, dashboard, içerik CRUD, video upload, encoding izleme
14. **UX & Responsive** → Skeleton'lar, error/empty state'ler, mobil uyum, animasyonlar
15. **Test & Doğrulama** → Uçtan uca tüm akışların tarayıcıda doğrulanması
