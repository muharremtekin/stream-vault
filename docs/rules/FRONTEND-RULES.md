# StreamVault — Frontend Kuralları

> **Bu dosya frontend projesinin anayasasıdır.**
> Buradaki kurallar tartışmaya açık değildir. Hiçbir gerekçeyle, hiçbir "hızlıca halledelim" bahanesiyle ihlal edilemez.
> Lint kuralları otomatik kontrol edilir. İhlal eden PR merge edilmez.

---
## 1. Proje Yapısı & Dosya İsimlendirme

### 1.1 Klasör Yapısı

```
web/src/
├── app/                  # Next.js App Router sayfa ve layout'ları
│   ├── (auth)/           # Auth layout grubu (login, register)
│   ├── (main)/           # Ana layout grubu (navbar + footer)
│   └── globals.css
├── components/           # React bileşenleri (UI, layout, domain)
│   ├── ui/               # Genel UI: button, input, modal, skeleton
│   ├── layout/           # Navbar, footer, sidebar
│   ├── auth/             # Login/register formları
│   ├── browse/           # Content card, hero banner, content row
│   ├── content/          # Film/dizi detay bileşenleri
│   ├── player/           # Video player bileşenleri
│   ├── search/           # Arama bileşenleri
│   ├── subscription/     # Abonelik bileşenleri
│   ├── notification/     # Bildirim bileşenleri
│   └── admin/            # Admin panel bileşenleri
├── lib/
│   ├── api/              # Axios API modülleri (auth.ts, catalog.ts, ...)
│   ├── hooks/            # Custom hook'lar (use-auth.ts, use-search.ts, ...)
│   ├── stores/           # Zustand store'ları (auth-store.ts, player-store.ts, ...)
│   ├── types/            # TypeScript tipleri (auth.ts, catalog.ts, common.ts, ...)
│   ├── validations/      # Zod schema'ları (login-schema.ts, ...)
│   └── utils/            # Utility fonksiyonlar (cn.ts, format.ts, constants.ts)
└── proxy.ts              # Auth redirect (Next.js 16 — middleware.ts DEĞİL)
```

### 1.2 Dosya İsimlendirme

```
Bileşenler:  kebab-case.tsx     (content-card.tsx, hero-banner.tsx)
Hook'lar:    use-kebab-case.ts  (use-auth.ts, use-search.ts)
Store'lar:   kebab-case.ts      (auth-store.ts, player-store.ts)
API/Tip/Util:kebab-case.ts      (catalog.ts, auth.ts, cn.ts, format.ts)
Validasyon:  kebab-case.ts      (login-schema.ts, register-schema.ts)
Klasörler:   kebab-case          Sayfalar: page.tsx / layout.tsx / error.tsx (Next.js)
```

### 1.3 Barrel Export Yasağı

Barrel export (`index.ts`) YASAKTIR — tree-shaking'i bozar, circular dependency riski oluşturur. Her dosya doğrudan import edilir: `import { Button } from '@/components/ui/button';`

---
## 2. TypeScript Kuralları

```
- strict: true ZORUNLU. tsconfig.json'da strict mode kapatılamaz.
- any tipi YASAKTIR. unknown veya spesifik tip kullan.
- as type casting YASAKTIR (DOM event casting hariç — yanına yorum ekle).
- ! (non-null assertion) YASAKTIR. Optional chaining (?.) veya null check kullan.
- enum YASAKTIR. as const obje kullan (tree-shaking dostu).
- interface: Props ve API response'ları için. type: Union, intersection, utility tipler için.
```

```typescript
// ❌ YANLIŞ — enum
enum VideoStatus { Ready = 'ready', Encoding = 'encoding', Failed = 'failed' }

// ✅ DOĞRU — as const
const VIDEO_STATUS = { Ready: 'ready', Encoding: 'encoding', Failed: 'failed' } as const;
type VideoStatus = (typeof VIDEO_STATUS)[keyof typeof VIDEO_STATUS];
```

```typescript
// ❌ YANLIŞ — any + non-null assertion
function parse(data: any) { return data.items; }
const user = store.getState().user!;

// ✅ DOĞRU — unknown + null check
function parse(data: unknown): Item[] {
  if (!isItemsResponse(data)) throw new Error('Invalid response');
  return data.items;
}
const user = store.getState().user;
if (!user) throw new Error('User not authenticated');
```

---
## 3. React & Next.js Kuralları

**Server Component tercih edilir.** `"use client"` sadece state, event handler veya browser API kullanıldığında eklenir.

```tsx
// ❌ YANLIŞ — Next.js 16'da senkron params (çalışmaz)
export default function MoviePage({ params }: { params: { id: string } }) {
  return <div>{params.id}</div>;
}

// ✅ DOĞRU — async params (Next.js 16 zorunlu)
export default async function MoviePage(props: { params: Promise<{ id: string }> }) {
  const { id } = await props.params;
  return <div>{id}</div>;
}
```

**Not:** `searchParams` de aynı şekilde async'tir: `const { q } = await props.searchParams;`

**Diğer Kurallar:**
- `middleware.ts` DEĞİL, `proxy.ts` kullanılır (Next.js 16 convention).
- Inline style YASAKTIR. Tek istisna: dinamik değerler (ör: `style={{ width: \`${progress}%\` }}`).
- Array render'da `index` key YASAKTIR. Benzersiz ID kullan.
- İç içe ternary (2+ seviye) YASAKTIR. Erken return kullan.
- Boolean prop'lar `is`, `has`, `should`, `can` ile başlar. Event prop'ları `on` ile başlar.

```tsx
// ❌ YANLIŞ — iç içe ternary
return isLoading ? <Skeleton /> : error ? <Error /> : data ? <Content /> : <Empty />;

// ✅ DOĞRU — erken return
if (isLoading) return <Skeleton />;
if (error) return <Error />;
if (!data) return <Empty />;
return <Content data={data} />;
```

---
## 4. Bileşen Kuralları

```
- Bir bileşen MAX 180 satır. Daha uzunsa parçala. İstisna yok.
- Tek sorumluluk. VideoPlayer hem video oynatıp hem yorum göstermez.
- Props interface: {BileşenAdı}Props formatı zorunlu.
- Prop drilling YASAKTIR (3+ seviye). Zustand veya Context kullan.
- React 19: forwardRef gereksiz — ref doğrudan prop olarak geçirilir.
```

```tsx
// ❌ YANLIŞ — isimsiz props interface
interface Props { title: string; year: number; }

// ✅ DOĞRU
interface ContentCardProps { title: string; year: number; }
```

```tsx
// ❌ forwardRef (React 19'da gereksiz)
const Input = forwardRef<HTMLInputElement, InputProps>((props, ref) => <input ref={ref} {...props} />);
// ✅ ref doğrudan prop
function Input({ ref, ...props }: InputProps & { ref?: React.Ref<HTMLInputElement> }) {
  return <input ref={ref} {...props} />;
}
```

---
## 5. State Management

### 5.1 Zustand Kuralları

- Domain başına tek store: `auth-store`, `player-store`, `notification-store`.
- Store'da business logic YASAKTIR. Sadece state + basit setter.
- `persist` middleware SADECE auth store için (token'lar).
- Store'dan tüm state'i çekme. **Selector** ile sadece gereken alanı al.

```typescript
// ❌ const store = useAuthStore(); return <span>{store.user?.email}</span>; // tüm store (gereksiz re-render)
// ✅ const email = useAuthStore((s) => s.user?.email); // selector
```

### 5.2 TanStack Query Kuralları

- Server state (film listesi, arama sonuçları) **ASLA** Zustand'da tutulmaz. TanStack Query kullanılır.
- Query key: `[domain, action, params]` formatı zorunlu.
- `staleTime` her query'de belirtilir (default 0 gereksiz refetch yapar).

```typescript
// ❌ queryKey: ['getMovieById', id]    // fiil kullanma
// ❌ queryKey: ['movies']              // domain yok
// ✅ queryKey: ['catalog', 'movie', id]
// ✅ queryKey: ['catalog', 'movies', { page, genre }]
// ✅ queryKey: ['search', 'results', { query, filters }]
// ✅ queryKey: ['streaming', 'progress', contentId]
```

---
## 6. API Client

- **Tek Axios instance** (`src/lib/api/client.ts`). Başka yerde instance oluşturmak YASAKTIR.
- Component içinde doğrudan `fetch` veya `axios` çağrısı YASAKTIR. Tüm çağrılar `src/lib/api/` altında.
- Request interceptor: `Authorization: Bearer {token}` + `X-Profile-Id` otomatik eklenir.
- Response interceptor: 401 → refresh token dene → başarısızsa logout + `/login` redirect.
- Her API fonksiyonu **dönüş tipini açıkça belirtir**. `any` dönüş tipi YASAKTIR.
- Default timeout: 10s. Video upload için özel timeout.

```typescript
// ❌ YANLIŞ — component içinde doğrudan API çağrısı
axios.get('/api/catalog/movies/123').then(res => setMovie(res.data));

// ✅ DOĞRU — src/lib/api/catalog.ts + TanStack Query
export async function getMovie(id: string): Promise<Movie> {
  const { data } = await client.get<Movie>(`/api/catalog/movies/${id}`);
  return data;
}
// component: useQuery({ queryKey: ['catalog', 'movie', id], queryFn: () => getMovie(id) })
```

---
## 7. Form & Validasyon

- Her form **React Hook Form + Zod** kullanır. İstisna yok.
- **Schema-first:** Önce Zod schema tanımlanır, tip `z.infer<typeof schema>` ile türetilir. Ayrı interface YASAKTIR.
- Hata gösterimi tutarlı: her input altında kırmızı hata metni.
- Submit sırasında buton `disabled` + loading state. Çift tıklama önlenir.

```typescript
// ✅ DOĞRU — schema-first
const loginSchema = z.object({
  email: z.string().email('Geçerli bir email adresi giriniz'),
  password: z.string().min(8, 'Şifre en az 8 karakter olmalıdır'),
});
type LoginFormData = z.infer<typeof loginSchema>;
```

---
## 8. Styling & Tailwind

```
- Inline style YASAKTIR (dinamik değerler hariç — yanına yorum ekle).
- CSS Modules YASAKTIR. Tüm stil Tailwind CSS ile yapılır.
- Koşullu class birleşimi için cn() utility (clsx + tailwind-merge) zorunlu.
- Sihirli sayılar YASAKTIR: w-[347px] yerine yakın standart değer (w-80, w-96) kullan.
- Responsive: mobile-first sıra → base → sm: → md: → lg: → xl:
- Renk kodları doğrudan yazılmaz. CSS variable'lar kullanılır (bg-background, text-foreground).
- Tailwind CSS 4: @import "tailwindcss" kullanılır (@tailwind base/components/utilities DEĞİL).
```

```tsx
// ❌ YANLIŞ — string concatenation
<button className={`px-4 py-2 ${isPrimary ? 'bg-primary' : 'bg-muted'} ${isDisabled ? 'opacity-50' : ''}`}>

// ✅ DOĞRU — cn() utility
<button className={cn('px-4 py-2', isPrimary ? 'bg-primary' : 'bg-muted', isDisabled && 'opacity-50')}>
```

---
## 9. Hata Yönetimi

- Her route grubu için `error.tsx` dosyası ZORUNLU.
- Hata **ASLA** yutulmaz. catch bloğu boş bırakılamaz.
- Kullanıcıya Sonner toast ile hata gösterilir.
- `console.log` / `console.error` production'da YASAKTIR.

```typescript
// ❌ YANLIŞ
try { await subscribe(planId); } catch { }

// ✅ DOĞRU
try {
  await subscribe(planId);
  toast.success('Abonelik başarıyla oluşturuldu');
} catch (error) {
  toast.error(getErrorMessage(error));
}
```

Hata mesajı çıkarma utility'si (`src/lib/utils/error.ts`) oluşturulur. Backend `ApiError` formatı parse edilir.

---
## 10. Performans

- HTML `<img>` YASAKTIR. Tüm görseller `next/image` ile optimize edilir (`width`/`height` veya `fill`).
- Ağır bileşenler (video player, modal) `next/dynamic` ile lazy load edilir (`ssr: false`).
- Wildcard import YASAKTIR: `import { Play } from 'lucide-react'` (tree-shaking dostu).
- Debounce: arama input'u 300ms. Throttle: scroll event. Progress save: 15s interval.

```tsx
// ❌ <img src="/poster.jpg" />
// ✅ <Image src="/poster.jpg" alt="Film posteri" width={300} height={450} />
```

---
## 11. Güvenlik

- `dangerouslySetInnerHTML` YASAKTIR. İstisna yok. Kullanıcı girdisi raw HTML olarak render edilmez.
- Client'a açık env değişkenleri `NEXT_PUBLIC_` prefix'i ile. Bu prefix'te **SECRET değer bulunmaz**.
- Token'lar localStorage'da (Zustand persist). XSS vektörleri minimize edilir.
- Tüm form input'ları Zod ile sanitize edilir (`trim()`, `max()`).
- Admin route'ları hem `proxy.ts`'de hem component'te role kontrolü yapar.
- `NEXT_PUBLIC_JWT_SECRET` gibi secret'lar ❌ YASAKTIR. Secret sadece server-side `.env`'de yaşar.

---
## 12. Test

- Framework: **Vitest + React Testing Library**. Jest kullanılmaz.
- Snapshot test YASAKTIR. Davranış testi yazılır.
- Test dosyası: `{bileşen-adı}.test.tsx` veya `{hook-adı}.test.ts`.
- Test adı: `{Ne} — {senaryo} — {beklenen sonuç}` formatı.
- Kullanıcı perspektifinden test: `getByRole`, `getByText` tercih edilir. `getByTestId` son çare.
- Her test izole çalışır. Testler arası state paylaşılmaz.

```typescript
// ❌ YANLIŞ — snapshot test
expect(container).toMatchSnapshot();

// ✅ DOĞRU — davranış testi (kullanıcı perspektifi)
it('LoginForm — geçersiz email — hata mesajı gösterir', async () => {
  render(<LoginForm />);
  await userEvent.type(screen.getByLabelText('Email'), 'geçersiz');
  await userEvent.click(screen.getByRole('button', { name: 'Giriş Yap' }));
  expect(screen.getByText('Geçerli bir email adresi giriniz')).toBeInTheDocument();
});
```

---
## 13. i18n (Çoklu Dil)

- Kütüphane: **next-intl**. Başka i18n kütüphanesi kullanılmaz.
- Desteklenen diller: `en` (default), `tr`. Default locale: `en`.
- Çeviri dosyaları: `messages/en.json`, `messages/tr.json`. Her iki dosya **aynı key yapısına** sahip olmalı.
- Hardcoded kullanıcı metni YASAKTIR. Tüm UI metinleri `useTranslations()` ile çekilir.
- Key namespace: sayfa/domain bazlı. `t('auth.loginTitle')`, `t('browse.heroPlay')`, `t('common.save')`.
- Component içinde string literal UI metni commit edilemez. Teknik string'ler (CSS class, API path) hariç.

```tsx
// ❌ YANLIŞ — hardcoded metin
<button>Giriş Yap</button>
<p>Film bulunamadı</p>

// ✅ DOĞRU — next-intl
import { useTranslations } from 'next-intl';

const t = useTranslations('auth');
<button>{t('login')}</button>
<p>{t('notFound')}</p>
```

```json
// messages/en.json                           // messages/tr.json
{ "auth": { "login": "Sign In" } }           { "auth": { "login": "Giriş Yap" } }
```

- Yeni UI metni → **her iki dil dosyası** aynı anda güncellenir. Eksik key YASAKTIR.
- Tarih/sayı: `next-intl` formatter'ları (`format.dateTime()`, `format.number()`).

---
## 14. Import Kuralları

**Sıra:** (1) React → (2) Next.js → (3) External → (4) Internal `@/` → (5) Types. Gruplar arası boş satır.

```typescript
import { useEffect } from 'react';       // 1) React
import Image from 'next/image';           // 2) Next.js
import { Play } from 'lucide-react';      // 3) External
import { cn } from '@/lib/utils/cn';      // 4) Internal @/
import type { Movie } from '@/lib/types/catalog'; // 5) Types
```

- `../../..` relative import (2+ seviye) YASAKTIR. `@/` path alias kullanılır.
- Tip import'ları `import type { X }` ile. Circular import YASAKTIR.

---
## 15. Erişilebilirlik (a11y)

- Semantik HTML: `<div onClick>` yerine `<button>` kullan. `<nav>`, `<main>`, `<article>` kullanılır.
- Her `<Image>` bileşeninde anlamlı `alt` text. Dekoratif görseller: `alt=""`.
- İkon-only butonlarda `aria-label` ZORUNLU.
- `outline-none` YASAKTIR. `focus-visible:ring-2` ile keyboard focus gösterilir.
- Modal açıkken focus trap + ESC ile kapatma.

```tsx
// ❌ YANLIŞ
<div onClick={handlePlay} className="cursor-pointer"><Play /></div>

// ✅ DOĞRU
<button onClick={handlePlay} aria-label="Videoyu oynat"><Play /></button>
```

---
## 16. Pre-PR Checklist

```
[ ] TypeScript hata vermiyor (tsc --noEmit)
[ ] ESLint hata vermiyor (eslint .)
[ ] "use client" sadece gereken bileşenlerde var
[ ] Yeni bileşen 150 satırı geçmiyor
[ ] Yeni sayfa için loading.tsx ve error.tsx mevcut
[ ] API çağrıları src/lib/api/ altında, component içinde fetch yok
[ ] Formlar React Hook Form + Zod kullanıyor
[ ] Görseller next/image ile optimize edilmiş
[ ] Tüm interaktif elemanlar keyboard ile erişilebilir
[ ] Responsive kontrol edildi (mobile, tablet, desktop)
[ ] Inline style kullanılmamış (dinamik değerler hariç)
[ ] any tipi kullanılmamış
[ ] Query key convention'a uygun
[ ] console.log / console.error temizlendi
[ ] Yeni UI metni varsa en.json ve tr.json aynı anda güncellendi
[ ] .env.local.example güncellendi (yeni variable varsa)
[ ] Commit mesajları formata uygun
```

---

> **Son söz:** Bu kurallar "güzel olurdu" değil "olmazsa olmaz"dır.
> Bir kural anlamsız geliyorsa tartışılabilir ve bu dosya güncellenebilir.
> Ama güncellenmediği sürece herkes bu kurallara uyar — insan da, agent da.
