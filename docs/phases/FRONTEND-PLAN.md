# StreamVault — Frontend: Web Uygulaması

> **Süre:** ~2-3 Hafta (backend fazlarıyla paralel veya sonrasında)
> **Teknoloji:** Next.js 14 (App Router) + TypeScript + Tailwind CSS
> **Hedef:** 8 backend servisinin tüm özelliklerini kullanan, Netflix görünümlü fonksiyonel bir web arayüzü.
> **Felsefe:** Abartısız ama tam. Her backend endpoint'inin karşılığı olan bir UI bileşeni. Görsel olarak tanınır bir streaming deneyimi.

---

## 1. Neden Next.js?

| Kriter | Karar | Sebep |
|--------|-------|-------|
| Framework | Next.js 14 (App Router) | SSR/SSG ile hızlı yükleme, API route'ları ile BFF pattern |
| Dil | TypeScript | Backend proto/DTO'larıyla uyumlu tip güvenliği |
| Styling | Tailwind CSS | Hızlı geliştirme, Netflix-dark tema için ideal |
| State | Zustand | Basit, boilerplate'siz global state |
| Data Fetching | TanStack Query (React Query) | Cache, retry, optimistic update |
| Video Player | hls.js | HLS adaptive bitrate playback |
| WebSocket | Native WebSocket + reconnect | Bildirimler için |
| Form | React Hook Form + Zod | Validasyonlu formlar |
| Icons | Lucide React | Hafif, tutarlı icon seti |
| Toast/Alert | Sonner | Bildirim toast'ları |

---

## 2. Proje Yapısı

```
streamvault/
├── ... (backend yapısı)
│
└── web/                                    # Next.js frontend
    ├── public/
    │   ├── logo.svg
    │   └── placeholder-poster.jpg
    │
    ├── src/
    │   ├── app/                            # Next.js App Router
    │   │   ├── layout.tsx                  # Root layout (dark tema, font, providers)
    │   │   ├── page.tsx                    # Landing page (login olmamışsa)
    │   │   ├── globals.css
    │   │   │
    │   │   ├── (auth)/                     # Auth layout grubu (navbar yok)
    │   │   │   ├── layout.tsx
    │   │   │   ├── login/
    │   │   │   │   └── page.tsx
    │   │   │   └── register/
    │   │   │       └── page.tsx
    │   │   │
    │   │   ├── (main)/                     # Ana uygulama layout grubu (navbar var)
    │   │   │   ├── layout.tsx              # Navbar + NotificationProvider
    │   │   │   ├── browse/
    │   │   │   │   └── page.tsx            # Ana sayfa (Netflix home)
    │   │   │   ├── search/
    │   │   │   │   └── page.tsx            # Arama sonuçları
    │   │   │   ├── movie/[id]/
    │   │   │   │   └── page.tsx            # Film detay
    │   │   │   ├── series/[id]/
    │   │   │   │   └── page.tsx            # Dizi detay
    │   │   │   ├── watch/[contentId]/
    │   │   │   │   └── page.tsx            # Video player sayfası
    │   │   │   ├── my-list/
    │   │   │   │   └── page.tsx            # Watchlist
    │   │   │   ├── genre/[slug]/
    │   │   │   │   └── page.tsx            # Tür bazlı listeleme
    │   │   │   │
    │   │   │   ├── account/
    │   │   │   │   ├── page.tsx            # Hesap genel bakış
    │   │   │   │   ├── profiles/
    │   │   │   │   │   └── page.tsx        # Profil yönetimi
    │   │   │   │   ├── subscription/
    │   │   │   │   │   └── page.tsx        # Abonelik yönetimi
    │   │   │   │   └── notifications/
    │   │   │   │       └── page.tsx        # Bildirim ayarları
    │   │   │   │
    │   │   │   └── admin/                  # Admin paneli
    │   │   │       ├── layout.tsx          # Admin guard
    │   │   │       ├── page.tsx            # Admin dashboard
    │   │   │       ├── content/
    │   │   │       │   ├── page.tsx        # İçerik listesi
    │   │   │       │   ├── new/
    │   │   │       │   │   └── page.tsx    # Yeni içerik ekle
    │   │   │       │   └── [id]/
    │   │   │       │       └── page.tsx    # İçerik düzenle + video upload
    │   │   │       └── encoding/
    │   │   │           └── page.tsx        # Encoding job listesi
    │   │   │
    │   │   └── api/                        # Next.js API routes (BFF)
    │   │       └── auth/
    │   │           └── [...nextauth]/
    │   │               └── route.ts        # Token yönetimi (httpOnly cookie)
    │   │
    │   ├── components/
    │   │   ├── ui/                         # Temel UI bileşenleri
    │   │   │   ├── button.tsx
    │   │   │   ├── input.tsx
    │   │   │   ├── modal.tsx
    │   │   │   ├── dropdown.tsx
    │   │   │   ├── skeleton.tsx            # Loading skeleton
    │   │   │   ├── badge.tsx
    │   │   │   ├── tooltip.tsx
    │   │   │   ├── progress-bar.tsx
    │   │   │   └── toast.tsx
    │   │   │
    │   │   ├── layout/                     # Layout bileşenleri
    │   │   │   ├── navbar.tsx              # Üst navigasyon
    │   │   │   ├── footer.tsx
    │   │   │   ├── sidebar.tsx             # Admin sidebar
    │   │   │   ├── profile-switcher.tsx    # Profil değiştirme dropdown
    │   │   │   └── notification-bell.tsx   # Bildirim ikonu + dropdown
    │   │   │
    │   │   ├── auth/
    │   │   │   ├── login-form.tsx
    │   │   │   ├── register-form.tsx
    │   │   │   └── profile-select.tsx      # "Kim izliyor?" ekranı
    │   │   │
    │   │   ├── browse/                     # Ana sayfa bileşenleri
    │   │   │   ├── hero-banner.tsx         # Büyük öne çıkan içerik
    │   │   │   ├── content-row.tsx         # Yatay kaydırmalı içerik sırası
    │   │   │   ├── content-card.tsx        # Tek içerik kartı (poster + hover)
    │   │   │   ├── content-card-hover.tsx  # Hover'da açılan detay kartı
    │   │   │   ├── genre-tags.tsx          # Tür etiketleri
    │   │   │   └── trending-badge.tsx      # "#1 Trend" rozeti
    │   │   │
    │   │   ├── content/                    # İçerik detay bileşenleri
    │   │   │   ├── content-hero.tsx        # Banner + bilgi + butonlar
    │   │   │   ├── content-info.tsx        # Açıklama, oyuncular, yönetmen
    │   │   │   ├── episode-list.tsx        # Dizi bölüm listesi
    │   │   │   ├── season-selector.tsx     # Sezon seçici
    │   │   │   ├── similar-content.tsx     # "Benzerleri" bölümü
    │   │   │   ├── rating-stars.tsx        # Puanlama yıldızları
    │   │   │   ├── maturity-badge.tsx      # Yaş sınıfı rozeti (PG-13, R)
    │   │   │   └── add-to-list-button.tsx  # Watchlist'e ekle butonu
    │   │   │
    │   │   ├── player/                     # Video player bileşenleri
    │   │   │   ├── video-player.tsx        # Ana player (hls.js wrapper)
    │   │   │   ├── player-controls.tsx     # Play/pause, seek, volume
    │   │   │   ├── quality-selector.tsx    # Kalite seçimi (360p, 720p...)
    │   │   │   ├── progress-tracker.tsx    # İzleme pozisyonu gönderme
    │   │   │   └── player-overlay.tsx      # Başlık, geri butonu
    │   │   │
    │   │   ├── search/
    │   │   │   ├── search-bar.tsx          # Arama çubuğu + autocomplete
    │   │   │   ├── search-results.tsx      # Sonuç grid'i
    │   │   │   ├── search-filters.tsx      # Facet filtreleri (tür, yıl, rating)
    │   │   │   └── search-highlight.tsx    # Vurgulu metin gösterimi
    │   │   │
    │   │   ├── subscription/
    │   │   │   ├── plan-cards.tsx          # 3 plan kartı (Basic, Standard, Premium)
    │   │   │   ├── payment-form.tsx        # Kart bilgisi formu (mock)
    │   │   │   ├── subscription-status.tsx # Aktif abonelik durumu
    │   │   │   └── invoice-table.tsx       # Fatura geçmişi tablosu
    │   │   │
    │   │   ├── notification/
    │   │   │   ├── notification-provider.tsx # WebSocket bağlantı yönetimi
    │   │   │   ├── notification-list.tsx    # Bildirim listesi
    │   │   │   ├── notification-item.tsx    # Tek bildirim kartı
    │   │   │   └── notification-prefs.tsx   # Tercih ayarları formu
    │   │   │
    │   │   └── admin/
    │   │       ├── content-form.tsx         # Film/dizi ekleme/düzenleme formu
    │   │       ├── video-uploader.tsx       # Drag & drop video upload
    │   │       ├── encoding-status.tsx      # Encoding pipeline durumu
    │   │       ├── encoding-progress.tsx    # İlerleme çubuğu (gerçek zamanlı)
    │   │       └── stats-cards.tsx          # Admin istatistik kartları
    │   │
    │   ├── lib/
    │   │   ├── api/                        # API client katmanı
    │   │   │   ├── client.ts               # Axios instance (base URL, interceptor)
    │   │   │   ├── auth.ts                 # register, login, refresh
    │   │   │   ├── catalog.ts              # movies, series, genres
    │   │   │   ├── streaming.ts            # manifest, progress, upload
    │   │   │   ├── search.ts               # search, autocomplete, trending
    │   │   │   ├── recommendation.ts       # recommendations, similar, home
    │   │   │   ├── subscription.ts         # plans, subscribe, cancel
    │   │   │   ├── notification.ts         # history, preferences, read
    │   │   │   └── encoding.ts             # jobs, status (admin)
    │   │   │
    │   │   ├── hooks/                      # Custom React hooks
    │   │   │   ├── use-auth.ts             # Auth state ve işlemleri
    │   │   │   ├── use-profile.ts          # Aktif profil yönetimi
    │   │   │   ├── use-search.ts           # Debounced arama + autocomplete
    │   │   │   ├── use-player.ts           # Player state (pozisyon, kalite)
    │   │   │   ├── use-watchlist.ts        # Watchlist CRUD
    │   │   │   ├── use-notifications.ts    # WebSocket + bildirim state
    │   │   │   └── use-subscription.ts     # Abonelik durumu
    │   │   │
    │   │   ├── stores/                     # Zustand stores
    │   │   │   ├── auth-store.ts           # User, token, profile
    │   │   │   ├── player-store.ts         # Player durumu
    │   │   │   └── notification-store.ts   # Bildirimler, unread count
    │   │   │
    │   │   ├── types/                      # TypeScript tipleri
    │   │   │   ├── auth.ts                 # User, Profile, LoginResponse
    │   │   │   ├── catalog.ts              # Movie, Series, Episode, Genre
    │   │   │   ├── streaming.ts            # StreamingInfo, Progress
    │   │   │   ├── search.ts               # SearchHit, Facets, Autocomplete
    │   │   │   ├── recommendation.ts       # RecommendedItem, HomeSection
    │   │   │   ├── subscription.ts         # Plan, Subscription, Payment
    │   │   │   ├── notification.ts         # Notification, Preferences
    │   │   │   └── common.ts               # PaginatedResponse, ApiError
    │   │   │
    │   │   └── utils/
    │   │       ├── format.ts               # Süre, tarih, para formatı
    │   │       ├── cn.ts                   # Tailwind class merge (clsx + twMerge)
    │   │       └── constants.ts            # API URL, renk kodları, vb.
    │   │
    │   └── middleware.ts                    # Auth redirect (login olmayanı yönlendir)
    │
    ├── tailwind.config.ts
    ├── next.config.js
    ├── tsconfig.json
    ├── package.json
    ├── Dockerfile
    └── .env.local.example
```

---

## 3. Sayfa Tasarımları ve Akışları

### 3.1 Landing Page (`/`)

```
┌─────────────────────────────────────────────────────────────────┐
│  [Logo: StreamVault]                          [Giriş Yap]      │
│                                                                 │
│                                                                 │
│              Film, dizi ve daha fazlası.                        │
│              Sınırsız izle. İstediğin zaman iptal et.           │
│                                                                 │
│              ┌──────────────────────────────────┐               │
│              │  Email adresinizi girin...   [Başla →]│           │
│              └──────────────────────────────────┘               │
│                                                                 │
│  ─────────────────────────────────────────────────────────      │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ 📺 TV'de    │  │ 📱 Her      │  │ 👤 Profil   │             │
│  │ izle        │  │ yerde izle  │  │ oluştur     │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
│                                                                 │
│  ─────────────────────────────────────────────────────────      │
│                                                                 │
│  Sıkça Sorulan Sorular (accordion)                              │
│  ├── StreamVault nedir?                                         │
│  ├── Aylık ücreti ne kadar?                                     │
│  ├── Nerede izleyebilirim?                                      │
│  └── Nasıl iptal ederim?                                        │
└─────────────────────────────────────────────────────────────────┘

Kullanılan API: Yok (statik sayfa)
Auth durumu: Giriş yapmış → /browse'a redirect
```

### 3.2 Login / Register (`/login`, `/register`)

```
┌─────────────────────────────────────────────────────────────────┐
│  [Logo: StreamVault]                                            │
│                                                                 │
│         ┌────────────────────────────────┐                      │
│         │                                │                      │
│         │         Giriş Yap              │                      │
│         │                                │                      │
│         │  ┌──────────────────────────┐  │                      │
│         │  │  Email                   │  │                      │
│         │  └──────────────────────────┘  │                      │
│         │  ┌──────────────────────────┐  │                      │
│         │  │  Şifre                   │  │                      │
│         │  └──────────────────────────┘  │                      │
│         │                                │                      │
│         │  [      Giriş Yap       ]      │                      │
│         │                                │                      │
│         │  Hesabın yok mu? Kayıt ol      │                      │
│         │                                │                      │
│         └────────────────────────────────┘                      │
│                                                                 │
│  (Arka plan: bulanık film posterleri collage)                   │
└─────────────────────────────────────────────────────────────────┘

Kullanılan API:
  POST /api/auth/login
  POST /api/auth/register
  POST /api/auth/refresh
```

### 3.3 Profil Seçimi (Login sonrası)

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│                     Kim izliyor?                                │
│                                                                 │
│    ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐             │
│    │ 😀     │  │ 😎     │  │ 🧒     │  │   +    │             │
│    │        │  │        │  │        │  │ Profil │             │
│    │  Ali   │  │  Ayşe  │  │Çocuklar│  │  Ekle  │             │
│    └────────┘  └────────┘  └────────┘  └────────┘             │
│                                                                 │
│                   [ Profili Yönet ]                              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

Kullanılan API:
  GET  /api/users/me/profiles
  POST /api/users/me/profiles
```

### 3.4 Ana Sayfa / Browse (`/browse`)

```
┌─────────────────────────────────────────────────────────────────┐
│ [Logo]  Ana Sayfa  Diziler  Filmler  Listem  🔍  🔔(3)  👤     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │                                                             │ │
│ │  HERO BANNER (rastgele öne çıkan içerik)                    │ │
│ │                                                             │ │
│ │  Interstellar                                               │ │
│ │  ⭐ 8.7  |  2014  |  PG-13  |  2s 49dk                     │ │
│ │  Dünya'nın geleceği tehlikede...                            │ │
│ │                                                             │ │
│ │  [ ▶ Oynat ]  [ + Listem ]  [ ℹ Detay ]                    │ │
│ │                                                             │ │
│ └─────────────────────────────────────────────────────────────┘ │
│                                                                 │
│  Kaldığın Yerden Devam Et                                       │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐  →              │
│  │▶ 20% │ │▶ 65% │ │▶ 40% │ │▶ 85% │ │▶ 10% │               │
│  │ prog █│ │prog █│ │prog █│ │prog █│ │prog █│               │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘               │
│                                                                 │
│  Sizin İçin Seçtiklerimiz                                       │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐  →              │
│  │poster│ │poster│ │poster│ │poster│ │poster│               │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘               │
│                                                                 │
│  Türkiye'de Trend                                               │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐  →              │
│  │#1    │ │#2    │ │#3    │ │#4    │ │#5    │               │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘               │
│                                                                 │
│  Interstellar İzlediğiniz İçin                                  │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐  →              │
│  │poster│ │poster│ │poster│ │poster│ │poster│               │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘               │
│                                                                 │
│  Bilim Kurgu                                                    │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐  →              │
│  │poster│ │poster│ │poster│ │poster│ │poster│               │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘               │
│                                                                 │
│  Yeni Eklenenler                                                │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐  →              │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘               │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

Kullanılan API:
  GET /api/recommendations/home              → Tüm section'lar
  GET /api/stream/continue-watching          → Kaldığın yerden devam et
  GET /api/search/trending?window=week       → Trend listesi
  GET /api/catalog/movies/{id}               → Hero banner detay
```

### 3.5 İçerik Detay (`/movie/[id]`, `/series/[id]`)

```
┌─────────────────────────────────────────────────────────────────┐
│ [Logo]  Ana Sayfa  Diziler  Filmler  Listem  🔍  🔔  👤        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │  (banner arka plan, gradient overlay)                       │ │
│ │                                                             │ │
│ │  Interstellar                                               │ │
│ │  ⭐ 8.7 (1240 oy)  |  2014  |  PG-13  |  2s 49dk          │ │
│ │                                                             │ │
│ │  [ ▶ Oynat ]  [ + Listem ]                                  │ │
│ └─────────────────────────────────────────────────────────────┘ │
│                                                                 │
│  Dünya'nın geleceği tehlikede ve bir grup kaşif,               │
│  insanlık için yeni bir yuva bulmak üzere bir                  │
│  solucan deliğinden geçerek...                                 │
│                                                                 │
│  Yönetmen: Christopher Nolan                                    │
│  Oyuncular: Matthew McConaughey, Anne Hathaway, Jessica Chastain│
│  Türler: Bilim Kurgu, Dram, Macera                              │
│                                                                 │
│  Puanınız:  ☆ ☆ ☆ ☆ ☆  (henüz puanlamadınız)                  │
│                                                                 │
│  ─── Dizi ise: Sezon ve Bölüm Listesi ───                      │
│                                                                 │
│  Sezon: [ 1 ▾ ]  [ 2 ]  [ 3 ]  [ 4 ]  [ 5 ]                   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ 1. Pilot                                    58 dk      │    │
│  │ Lise kimya öğretmeni Walter White...        [▶ Oynat]  │    │
│  ├─────────────────────────────────────────────────────────┤    │
│  │ 2. Cat's in the Bag                         48 dk      │    │
│  │ Walt ve Jesse...                            [▶ Oynat]  │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ─── Benzerleri ───                                             │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐  →              │
│  │poster│ │poster│ │poster│ │poster│ │poster│               │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘               │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

Kullanılan API:
  GET /api/catalog/movies/{id}                → Film detay
  GET /api/catalog/series/{id}                → Dizi detay + sezonlar
  GET /api/catalog/movies/{id}/streaming-info → Video durumu
  GET /api/recommendations/similar/{id}       → Benzerleri
  GET /api/stream/{id}/progress               → Kaldığı yer
  POST /api/users/me/ratings                  → Puanlama
  POST /api/users/me/profiles/{pid}/watchlist → Listeye ekle
```

### 3.6 Video Player (`/watch/[contentId]`)

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│  ← Geri                              Interstellar               │
│                                                                 │
│                                                                 │
│                                                                 │
│                          advancement                                     │
│                      ▶ VIDEO ALANI                              │
│                       (hls.js)                                  │
│                                                                 │
│                                                                 │
│                                                                 │
│                                                                 │
│  ▐▐   advancement━━━━━━━━━●━━━━━━━━━━━━━━━  1:23:45 / 2:49:00   │
│  🔊▁▂▃▄  |  720p ▾  |  ⛶ Tam Ekran                             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

Tam ekran, kontroller 3 saniye hareketsizlikte gizlenir.

Video player davranışı:
  1. Sayfa yüklendiğinde:
     → GET /stream/{id}/progress (kaldığı yeri öğren)
     → GET /stream/{id}/manifest.m3u8 (HLS manifest al)
     → hls.js ile video yükleme başlat

  2. Oynatma sırasında:
     → hls.js adaptive bitrate otomatik kalite seçer
     → Kullanıcı manuel kalite seçebilir (quality-selector)
     → Her 15 saniyede: POST /api/stream/{id}/progress

  3. Kalite seçici:
     → Otomatik (önerilen)
     → 360p  (Basic+)
     → 720p  (Basic+)
     → 1080p (Standard+)
     → 4K    (Premium)   ← Tier'e göre disabled

  4. Oynatma bittiğinde:
     → Sonraki bölüm önerisi (dizi ise)
     → "Benzerleri" önerisi (film ise)

Kullanılan API:
  GET  /stream/{id}/manifest.m3u8           → HLS master playlist
  GET  /stream/{id}/{quality}/playlist.m3u8 → Kalite playlist
  GET  /stream/{id}/{quality}/segment_*.ts  → Video segmentleri
  GET  /api/stream/{id}/progress            → Kaldığı yer
  POST /api/stream/{id}/progress            → Pozisyon kaydet
```

### 3.7 Arama (`/search`)

```
┌─────────────────────────────────────────────────────────────────┐
│ [Logo]  Ana Sayfa  Diziler  Filmler  Listem                     │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  🔍  interst|                                            │   │
│  ├──────────────────────────────────────────────────────────┤   │
│  │  🎬  Interstellar (2014)                                 │   │
│  │  📺  The Internship (2013)                               │   │
│  │  📺  Interlude in Prague (2017)                          │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ─── "interstellar" için 15 sonuç (12ms) ───                   │
│                                                                 │
│  Filtreler:                                                     │
│  Tür:    [Tümü ▾] [Bilim Kurgu (8)] [Dram (12)] [Aksiyon (6)] │
│  Yıl:    [2020+] [2015-2020] [2010-2015] [Daha Eski]          │
│  Rating: [⭐ 7+] [⭐ 8+] [⭐ 9+]                              │
│  Tür:    [Filmler] [Diziler] [Tümü]                            │
│  Sırala: [İlgililik ▾] [Puan] [Yıl] [İsim]                   │
│                                                                 │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐                          │
│  │poster│ │poster│ │poster│ │poster│                          │
│  │Inter-│ │The   │ │Gravi-│ │Arriv-│                          │
│  │stella│ │Marti-│ │ty    │ │al    │                          │
│  │⭐ 8.7│ │⭐ 8.0│ │⭐ 7.7│ │⭐ 7.9│                          │
│  └──────┘ └──────┘ └──────┘ └──────┘                          │
│                                                                 │
│  [ 1 ] [ 2 ] [ 3 ] ... [ 8 ]   Sayfalama                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

Kullanılan API:
  GET /api/search/autocomplete?q=inter       → Otomatik tamamlama
  GET /api/search?q=interstellar&genres=...  → Full-text arama + facet'ler
  GET /api/search/trending                   → Boş arama → trendler göster
```

### 3.8 Abonelik Yönetimi (`/account/subscription`)

```
┌─────────────────────────────────────────────────────────────────┐
│ [Logo]  Ana Sayfa  Diziler  Filmler  Listem  🔍  🔔  👤        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Abonelik Yönetimi                                              │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Mevcut Plan: Standard             Durum: ✅ Aktif       │   │
│  │  Aylık: ₺79.99                     Yenileme: 15 Mar 2026│   │
│  │  Özellikler: Full HD, 2 Ekran                            │   │
│  │                                                          │   │
│  │  [Plan Değiştir]   [İptal Et]                            │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ─── Planlar ───                                                │
│                                                                 │
│  ┌─────────────┐  ┌─────────────────┐  ┌─────────────────┐    │
│  │   Basic     │  │ ✓ Standard      │  │   Premium       │    │
│  │             │  │   (Mevcut)      │  │                 │    │
│  │  ₺49.99/ay │  │  ₺79.99/ay      │  │  ₺119.99/ay    │    │
│  │             │  │                 │  │                 │    │
│  │  720p       │  │  1080p          │  │  4K + Atmos     │    │
│  │  1 Ekran    │  │  2 Ekran        │  │  4 Ekran        │    │
│  │  Reklamsız  │  │  Reklamsız      │  │  Reklamsız      │    │
│  │  İndirme    │  │  İndirme        │  │  İndirme        │    │
│  │             │  │  Full HD        │  │  Ultra HD       │    │
│  │             │  │                 │  │  Dolby Atmos    │    │
│  │             │  │                 │  │                 │    │
│  │ [Downgrade] │  │   Mevcut Plan   │  │  [Upgrade]      │    │
│  └─────────────┘  └─────────────────┘  └─────────────────┘    │
│                                                                 │
│  ─── Fatura Geçmişi ───                                        │
│                                                                 │
│  Tarih         Plan       Tutar     Durum                       │
│  13 Şub 2026   Standard   ₺79.99    ✅ Ödendi                  │
│  13 Oca 2026   Basic      ₺49.99    ✅ Ödendi                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

Kullanılan API:
  GET  /api/plans                        → Plan listesi
  GET  /api/subscriptions/me             → Aktif abonelik
  POST /api/subscriptions                → Yeni abonelik
  PUT  /api/subscriptions/me/plan        → Plan değiştir
  POST /api/subscriptions/me/cancel      → İptal
  GET  /api/subscriptions/me/invoices    → Fatura geçmişi
```

### 3.9 Admin Panel (`/admin`)

```
┌─────────────────────────────────────────────────────────────────┐
│ [Logo]  ← Siteye Dön                      Admin: ali@test.com  │
├────────────┬────────────────────────────────────────────────────┤
│            │                                                    │
│  Dashboard │  Dashboard                                         │
│  İçerikler │                                                    │
│  Encoding  │  ┌──────────┐ ┌──────────┐ ┌──────────┐          │
│            │  │ 42       │ │ 3        │ │ 1,247    │          │
│            │  │ Toplam   │ │ Encoding │ │ Aktif    │          │
│            │  │ İçerik   │ │ İşlemde  │ │ Abone    │          │
│            │  └──────────┘ └──────────┘ └──────────┘          │
│            │                                                    │
│            │  Son Encoding İşlemleri                             │
│            │  ┌──────────────────────────────────────────────┐  │
│            │  │ Interstellar    1080p   ████████░░ 78%  ▶️  │  │
│            │  │ The Matrix      720p    ██████████ 100% ✅  │  │
│            │  │ Inception       4K      ██░░░░░░░░ 15%  ▶️  │  │
│            │  └──────────────────────────────────────────────┘  │
│            │                                                    │
├────────────┤────────────────────────────────────────────────────┤
│            │                                                    │
│  İçerikler │  İçerik Yönetimi           [+ Yeni Ekle]          │
│  (selected)│                                                    │
│            │  🔍 İçerik ara...                                  │
│            │                                                    │
│            │  Başlık           Tür       Video     Tarih        │
│            │  Interstellar     Film      ✅ Ready   13 Şub      │
│            │  Breaking Bad     Dizi      ✅ Ready   12 Şub      │
│            │  New Movie        Film      ⏳ Encoding 13 Şub     │
│            │  Draft Movie      Film      ❌ No Video 10 Şub     │
│            │                                                    │
├────────────┤────────────────────────────────────────────────────┤
│            │                                                    │
│  İçerik    │  Film Ekle / Düzenle                               │
│  Düzenleme │                                                    │
│            │  Başlık:      [                         ]          │
│            │  Açıklama:    [                         ]          │
│            │  Yıl:         [      ]  Rating: [ PG13 ▾]          │
│            │  Türler:      [Bilim Kurgu] [Dram] [+ Ekle]       │
│            │  Yönetmen:    [                         ]          │
│            │  Oyuncular:   [                    ] [+ Ekle]      │
│            │                                                    │
│            │  ─── Video Yükleme ───                             │
│            │  ┌──────────────────────────────────────────────┐  │
│            │  │                                              │  │
│            │  │     📁 Video dosyasını buraya sürükleyin     │  │
│            │  │        veya tıklayarak seçin                 │  │
│            │  │                                              │  │
│            │  │     MP4, MKV, AVI — Max 10GB                 │  │
│            │  │                                              │  │
│            │  └──────────────────────────────────────────────┘  │
│            │                                                    │
│            │  Upload durumu: ████████░░░░ 67% (342MB/512MB)    │
│            │                                                    │
│            │  [Kaydet]  [İptal]                                 │
│            │                                                    │
└────────────┴────────────────────────────────────────────────────┘

Kullanılan API:
  GET  /api/catalog/movies                 → İçerik listesi
  POST /api/catalog/movies                 → Yeni film ekle
  POST /api/catalog/series                 → Yeni dizi ekle
  POST /api/stream/upload                  → Video upload
  GET  /api/encoding/jobs                  → Encoding iş listesi
  GET  /api/encoding/jobs/{id}             → Encoding detay
```

---

## 4. Temel Bileşen Detayları

### 4.1 API Client (`lib/api/client.ts`)

```typescript
// Axios instance with interceptors

const client = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  timeout: 10000,
});

// Request interceptor: JWT token ekle
client.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  // Aktif profil ID'sini header olarak ekle
  const profileId = useAuthStore.getState().activeProfileId;
  if (profileId) {
    config.headers['X-Profile-Id'] = profileId;
  }
  return config;
});

// Response interceptor: 401 → refresh token dene
client.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      const refreshed = await tryRefreshToken();
      if (refreshed) {
        return client.request(error.config); // Retry
      }
      // Refresh de başarısız → logout
      useAuthStore.getState().logout();
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

### 4.2 Video Player (`components/player/video-player.tsx`)

```typescript
// hls.js entegrasyonu

// Temel akış:
// 1. manifest URL'i prop olarak al
// 2. hls.js instance oluştur, video elementine attach et
// 3. Quality level change event'lerini dinle
// 4. Error handling (network, media, fatal/non-fatal)
// 5. Cleanup: unmount'ta hls.destroy()

// Kalite seçimi:
// hls.currentLevel = index  → Manuel kalite
// hls.currentLevel = -1     → Otomatik (ABR)

// İzleme pozisyonu:
// useEffect ile 15 saniyede bir POST /api/stream/{id}/progress
// Sayfa kapanırken (beforeunload) son pozisyonu kaydet
// Sayfa açılırken GET /api/stream/{id}/progress ile kaldığı yere seek
```

### 4.3 WebSocket Notification Provider

```typescript
// Tüm uygulamayı saran provider

// Akış:
// 1. Kullanıcı login olduğunda WebSocket bağlantısı aç
//    ws://localhost:8080/ws/notifications?token=eyJ...
// 2. Gelen mesajları notification-store'a ekle
// 3. Toast göster (sonner ile)
// 4. Bağlantı koparsa exponential backoff ile yeniden bağlan
//    (1s → 2s → 4s → 8s → max 30s)
// 5. Logout'ta bağlantıyı kapat

// Heartbeat:
// Her 30 saniyede bir ping gönder
// 10 saniye içinde pong gelmezse reconnect
```

### 4.4 Content Card Hover Efekti

```
Normal durum:        Hover durumu (300ms delay):
┌──────────┐        ┌────────────────────┐
│          │        │                    │
│  poster  │   →    │    poster (zoom)   │
│          │        │                    │
│  Title   │        │  Title             │
└──────────┘        │  ⭐ 8.7 | 2014     │
                    │  Bilim Kurgu, Dram │
                    │  2s 49dk           │
                    │                    │
                    │  [▶] [+] [👍] [ℹ] │
                    └────────────────────┘

Hover kartı orijinal kartın üstüne scale(1.3) ile büyür.
z-index ile diğer kartların üstüne çıkar.
Kenarda ise sola/sağa kaydırılır (viewport dışına taşmaz).
```

---

## 5. State Management

### 5.1 Auth Store (Zustand)

```typescript
interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  activeProfile: Profile | null;
  profiles: Profile[];
  subscriptionTier: 'Basic' | 'Standard' | 'Premium' | null;

  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => void;
  refreshAuth: () => Promise<boolean>;
  setActiveProfile: (profile: Profile) => void;
  fetchProfiles: () => Promise<void>;
}
```

### 5.2 Notification Store

```typescript
interface NotificationState {
  notifications: Notification[];
  unreadCount: number;
  wsConnected: boolean;

  connect: (token: string) => void;
  disconnect: () => void;
  markAsRead: (id: string) => Promise<void>;
  markAllAsRead: () => Promise<void>;
  fetchHistory: (page: number) => Promise<void>;
}
```

### 5.3 Player Store

```typescript
interface PlayerState {
  contentId: string | null;
  isPlaying: boolean;
  currentTime: number;
  duration: number;
  currentQuality: string;
  availableQualities: QualityOption[];
  volume: number;
  isFullscreen: boolean;
  isBuffering: boolean;

  setQuality: (quality: string) => void;
  togglePlay: () => void;
  seek: (time: number) => void;
  setVolume: (vol: number) => void;
}
```

---

## 6. API ↔ Sayfa Eşleme (Tam Liste)

Hangi sayfada hangi backend API'si kullanılıyor:

```
Sayfa                     API Endpoint'leri
─────────────────────     ──────────────────────────────────────
/                         (statik — API yok)
/login                    POST /api/auth/login
/register                 POST /api/auth/register
(profil seçimi)           GET /api/users/me/profiles
                          POST /api/users/me/profiles
/browse                   GET /api/recommendations/home
                          GET /api/stream/continue-watching
                          GET /api/search/trending
/movie/[id]               GET /api/catalog/movies/{id}
                          GET /api/catalog/movies/{id}/streaming-info
                          GET /api/recommendations/similar/{id}
                          GET /api/stream/{id}/progress
                          POST /api/users/me/ratings
                          POST /api/users/me/profiles/{pid}/watchlist
/series/[id]              GET /api/catalog/series/{id}
                          GET /api/recommendations/similar/{id}
                          (aynı diğer endpoint'ler)
/watch/[contentId]        GET /stream/{id}/manifest.m3u8
                          GET /stream/{id}/{q}/playlist.m3u8
                          GET /stream/{id}/{q}/segment_*.ts
                          GET /api/stream/{id}/progress
                          POST /api/stream/{id}/progress
/search                   GET /api/search?q=...&genres=...
                          GET /api/search/autocomplete?q=...
                          GET /api/search/trending
/my-list                  GET /api/users/me/profiles/{pid}/watchlist
                          DELETE /api/users/me/profiles/{pid}/watchlist/{id}
/genre/[slug]             GET /api/catalog/genres/{slug}/content
/account                  GET /api/users/me
/account/profiles         GET /api/users/me/profiles
                          POST /api/users/me/profiles
/account/subscription     GET /api/plans
                          GET /api/subscriptions/me
                          POST /api/subscriptions
                          PUT /api/subscriptions/me/plan
                          POST /api/subscriptions/me/cancel
                          GET /api/subscriptions/me/invoices
/account/notifications    GET /api/notifications/preferences
                          PUT /api/notifications/preferences
(Navbar — her yerde)      WebSocket /ws/notifications
                          GET /api/notifications?unreadOnly=true
                          POST /api/notifications/{id}/read
                          POST /api/notifications/read-all
/admin                    GET /api/encoding/jobs
/admin/content            GET /api/catalog/movies
                          GET /api/catalog/series
/admin/content/new        POST /api/catalog/movies
                          POST /api/catalog/series
/admin/content/[id]       GET /api/catalog/movies/{id}
                          POST /api/stream/upload
/admin/encoding           GET /api/encoding/jobs
                          GET /api/encoding/jobs/{id}
```

---

## 7. Docker ve Gateway Entegrasyonu

### 7.1 Dockerfile

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/package*.json ./
COPY --from=builder /app/node_modules ./node_modules
EXPOSE 3001
CMD ["npm", "start"]
```

### 7.2 Docker Compose Eklentisi

```yaml
  web:
    build: ./web
    ports: ["3001:3001"]
    environment:
      - NEXT_PUBLIC_API_URL=http://localhost:8080
      - NEXT_PUBLIC_WS_URL=ws://localhost:8080
    depends_on:
      - gateway
```

### 7.3 Ağ Akışı

```
Browser (:3001)
    │
    ├── Sayfa yükleme (SSR/CSR) ──► Next.js (:3001)
    │
    └── API istekleri ──► API Gateway (:8080) ──► Backend Servisleri
        ├── REST (fetch/axios)
        ├── HLS (hls.js → video segment fetch)
        └── WebSocket (bildirimler)

Not: Frontend doğrudan backend servislerine ulaşmaz,
     her şey Gateway üzerinden geçer.
```

---

## 8. Haftalık İlerleme Planı

### Hafta 1: Temel + Auth + Browse

| Gün | Görev | Detay |
|-----|-------|-------|
| 1 | Proje kurulumu | Next.js + TypeScript + Tailwind. Layout, dark tema, font. API client (axios + interceptor). Zustand store'lar. Dockerfile. |
| 2 | Auth sayfaları | Login, register formları (React Hook Form + Zod). JWT token yönetimi. Auth middleware (korumalı route'lar). |
| 3 | Profil seçimi + Navbar | "Kim izliyor?" ekranı, profil switcher. Navbar (logo, navigation, arama ikonu, bildirim, profil). |
| 4 | Browse — Ana sayfa | Hero banner, content-row (yatay kaydırmalı), content-card. Recommendation API entegrasyonu. Continue watching. |
| 5 | Content card hover + Detay sayfası | Hover efekti (scale, bilgi gösterimi). Film/dizi detay sayfası. Puanlama. Watchlist ekleme. Sezon/bölüm listesi. |

### Hafta 2: Player + Search + Subscription

| Gün | Görev | Detay |
|-----|-------|-------|
| 1 | Video player | hls.js entegrasyonu. Player kontrolleri (play/pause, seek, volume, fullscreen). Kalite seçici. Keyboard shortcuts (space, f, ←, →). |
| 2 | Player — Progress + UX | İzleme pozisyonu kaydetme (15s interval). Kaldığın yerden devam. Buffering göstergesi. Oynatma bittiğinde öneri. |
| 3 | Arama | Search bar + debounced autocomplete. Arama sonuçları sayfası. Facet filtreleri (tür, yıl, rating). Highlight gösterimi. |
| 4 | Abonelik | Plan kartları. Ödeme formu (mock kart). Abonelik durumu. Plan değiştirme. İptal akışı. Fatura geçmişi tablosu. |
| 5 | Watchlist + Genre | Listem sayfası (grid, silme). Tür sayfası (/genre/bilim-kurgu). Sayfalama. |

### Hafta 3: Admin + Notifications + Polish

| Gün | Görev | Detay |
|-----|-------|-------|
| 1 | Admin — İçerik yönetimi | Admin layout + guard. İçerik listesi tablosu. Film/dizi ekleme formu. |
| 2 | Admin — Video upload + Encoding | Drag & drop upload bileşeni. Upload progress. Encoding job listesi. Gerçek zamanlı ilerleme çubuğu (polling). |
| 3 | Bildirimler | WebSocket provider. Notification bell + dropdown. Bildirim listesi sayfası. Okundu işaretleme. Toast bildirimleri. Tercih ayarları. |
| 4 | Responsive + UX polish | Mobil uyum (navbar hamburger, card boyutları). Loading skeleton'lar. Error state'ler. Empty state'ler. Page transition animasyonları. |
| 5 | Test + Dokümantasyon | Uçtan uca test (tüm akışları browser'da geç). README. Screenshot'lar. Bilinen sınırlamalar. |

---

## 9. Bitiş Kriterleri (Definition of Done)

### Auth ve Profil
- [ ] Kayıt ve giriş çalışıyor
- [ ] JWT token refresh otomatik çalışıyor (401 → retry)
- [ ] "Kim izliyor?" ekranı profilleri gösteriyor
- [ ] Yeni profil oluşturulabiliyor (max 5 sınırı UI'da gösteriliyor)
- [ ] Profil değiştirme çalışıyor

### Browse ve İçerik
- [ ] Ana sayfa tüm section'ları gösteriyor (kişisel, trend, tür bazlı)
- [ ] "Kaldığın Yerden Devam Et" sırası çalışıyor
- [ ] Content card hover efekti çalışıyor
- [ ] Film detay sayfası tüm bilgileri gösteriyor
- [ ] Dizi detay sayfası sezon/bölüm listesi gösteriyor
- [ ] Puanlama çalışıyor
- [ ] Watchlist ekleme/çıkarma çalışıyor
- [ ] Benzer içerikler gösteriliyor

### Video Player
- [ ] HLS video oynatılabiliyor (hls.js)
- [ ] Adaptive bitrate çalışıyor
- [ ] Manuel kalite seçimi çalışıyor
- [ ] Oynatma kontrolleri çalışıyor (play/pause, seek, volume, fullscreen)
- [ ] İzleme pozisyonu kaydediliyor (15s interval)
- [ ] Kaldığın yerden devam ediyor
- [ ] Keyboard shortcut'lar çalışıyor

### Arama
- [ ] Autocomplete 2+ karakterde öneri sunuyor
- [ ] Full-text arama sonuçları dönüyor
- [ ] Facet filtreleri çalışıyor (tür, yıl, rating)
- [ ] Arama sonuçlarında highlight görünüyor
- [ ] Boş aramada trendler gösteriliyor

### Abonelik
- [ ] 3 plan kartı görüntüleniyor
- [ ] Abonelik oluşturma (mock ödeme) çalışıyor
- [ ] Aktif abonelik durumu gösteriliyor
- [ ] Plan değiştirme çalışıyor
- [ ] İptal akışı çalışıyor
- [ ] Fatura geçmişi görüntüleniyor

### Bildirimler
- [ ] WebSocket bağlantısı kuruluyor
- [ ] Gerçek zamanlı bildirim alınıyor ve toast gösteriliyor
- [ ] Bildirim bell'inde unread sayısı görünüyor
- [ ] Bildirim listesi açılıyor
- [ ] Okundu işaretleme çalışıyor

### Admin
- [ ] Admin paneline sadece admin rolü erişebiliyor
- [ ] İçerik listesi görüntüleniyor
- [ ] Yeni film/dizi eklenebiliyor
- [ ] Video upload çalışıyor (progress göstergesiyle)
- [ ] Encoding job durumu görüntüleniyor

### Genel
- [ ] Dark tema tutarlı
- [ ] Loading skeleton'lar mevcut
- [ ] Error state'ler kullanıcı dostu
- [ ] `docker compose up` ile frontend ayağa kalkıyor
- [ ] Tüm API endpoint'leri kullanılıyor (boşta kalan yok)
