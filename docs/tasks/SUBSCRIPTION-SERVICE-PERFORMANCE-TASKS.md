# StreamVault - Subscription Service Performans Iyilestirme Gorevleri

> **Durum:** `services/subscription-service` kod taramasi sonucu tespit edilen performans darboğazlari icin duzeltme gorevleri.
> **Oncelik:** KRITIK > YUKSEK > ORTA
> **Tarih:** 2026-04-11

---

## 1. [KRITIK] Write Amplification ve Transaction Sinirlari

### Gorev 1.1: Repository seviyesindeki otomatik `SaveChangesAsync` cagri desenini kaldir
- [x] `SubscriptionRepository`, `PaymentRepository`, `InvoiceRepository`, `OutboxRepository`, `SagaRepository` icindeki her `AddAsync` ve `UpdateAsync` sonrasinda calisan otomatik `SaveChangesAsync` cagrilarini tespit et
- [x] Repository arayuzlerini ayni is akisi icinde birden fazla entity degisikligini biriktirebilecek sekilde yeniden tasarla
- [x] `SubscriptionDbContext` uzerinden tekil `SaveChangesAsync` veya unit-of-work benzeri bir commit noktasi belirle
- [x] Bu degisikligin `CreateSubscription`, `ChangePlan`, `CancelSubscription`, renewal worker ve compensation akislari ile uyumlu oldugunu dogrula

### Gorev 1.2: Saga adimlarini daha az roundtrip ile calisacak sekilde duzenle
- [x] `SubscriptionSaga` icindeki saga state guncellemelerini gereksiz ara commitler olmadan grupla
- [x] `ChangePlanSaga` icindeki saga state, payment, subscription ve outbox yazimlarini mantiksal fazlara ayir
- [x] Dis sistem cagri yapilan adimlar ile yalnizca veritabani iceren adimlar arasinda net transaction sinirlari tanimla
- [x] Basarili create ve change-plan akislarinda toplam veritabani write roundtrip sayisini olculebilir sekilde azalt

### Gorev 1.3: Kritik write akislari icin regresyon testleri ekle
- [x] Unit veya integration test seviyesinde saga adimlarinin dogru sirada commit edildigini dogrula
- [x] Basarisiz odeme ve compensation senaryolarinda veri tutarliliginin korundugunu test et
- [x] Outbox kayitlarinin ilgili domain degisikligi ile ayni commit fazinda olustugunu test et

**Kabul Kriterleri**
- [x] Basarili `CreateSubscription` akisinda mevcut tasarima gore belirgin sekilde daha az `SaveChangesAsync` cagrisi yapiliyor
- [x] Basarili `ChangePlan` akisinda gereksiz ara write operasyonlari kaldirildi
- [x] Veri tutarliligi bozulmadi; mevcut davranis ve hata senaryolari korunuyor

---

## 2. [KRITIK] Outbox Polling ve Batch Starvation Duzeltmeleri

### Gorev 2.1: Outbox retry/backoff mantigini SQL tarafina tasima
- [x] `outbox_messages` tablosuna `next_attempt_at` benzeri bir alan ekle
- [x] Retry alan kayitlarda sonraki deneme zamanini `IncrementRetryAsync` sirasinda hesapla ve yaz
- [x] `GetUnprocessedAsync` sorgusunu sadece islenmeye uygun kayitlari dondurecek sekilde guncelle
- [ ] Bellekte `RetryCount` ve `LastAttemptedAt` ile skip yapan mantigi sadeleştir

### Gorev 2.2: Outbox icin polling desenine uygun indeksler ekle
- [x] `processed_at`, `is_dead_letter`, `next_attempt_at`, `created_at` alanlarini kapsayan uygun composite veya partial index stratejisi belirle
- [ ] En sık calisan sorgu paternleri icin migration hazirla
- [ ] Index sonrasi sorgu planlarini `EXPLAIN` veya benzeri olcumle dogrula

### Gorev 2.3: Outbox processor throughput iyilestirmesi
- [x] Batch icindeki mesajalari hazirlik durumuna gore daha verimli isle
- [ ] Gerekirse publish + mark-processed akisini sinirli paralellik veya daha buyuk efektif batch mantigi ile iyilestir
- [x] Shutdown drain sirasinda da ayni due-message mantiginin korundugunu kontrol et

**Kabul Kriterleri**
- [ ] Backoff bekleyen kayitlar yeni hazir mesajlari bloke etmiyor
- [ ] Outbox sorgusu polling desenine uygun index kullanabiliyor
- [ ] Basarisiz publish sonrasinda retry davranisi islevsel olarak korunuyor

---

## 3. [YUKSEK] Subscription Renewal Worker Batch ve Set-Based Isleme

### Gorev 3.1: Renewal worker icin batch boyutu ve sirali isleme ekle
- [x] `GetExpiredSubscriptionsAsync` sorgusuna `batchSize` ve deterministik siralama ekle
- [x] `SubscriptionRenewalService` icinde tum kayitlari tek seferde cekmek yerine parcali isleme mantigi kur
- [x] Gerekirse configurable worker options ekleyerek interval ve batch boyutunu config uzerinden yonet

### Gorev 3.2: Renewal sorgusundaki gereksiz entity yuklemelerini kaldir
- [x] `GetExpiredSubscriptionsAsync` icindeki gereksiz `Include(s => s.Plan)` kullanimini kaldir
- [x] Sadece renewal icin gereken alanlari dondurecek projection veya hafif sorgu modeli kullan
- [x] Read-only path icin `AsNoTracking` kullaniminin uygun olup olmadigini degerlendir

### Gorev 3.3: Row-by-row update maliyetini azalt
- [x] Auto-renew kapali aboneliklerin `Expired` yapilmasi icin set-based update kullaniminin uygunlugunu degerlendir
- [x] Auto-renew acik aboneliklerde update + outbox ekleme akisini tek commit fazina yaklastir
- [ ] Renewal isleminde her kayit icin ayri write patlamasi olusmadigini olc

**Kabul Kriterleri**
- [ ] Renewal worker backlog buyudugunde tum satirlari bellekte tutmuyor
- [ ] Renewal sorgusu gereksiz `Plan` join'i yapmiyor
- [ ] Saatlik calismalarda DB ve memory baskisi azaltilmis durumda

---

## 4. [YUKSEK] Invoice Number Uretim Stratejisi

### Gorev 4.1: `CountAsync` tabanli invoice number uremini kaldir
- [x] `GenerateInvoiceNumberAsync` icinde gunluk kayit sayisi sayan yapinin performans ve race condition riskini kaldir
- [x] PostgreSQL sequence, counter tablosu veya DB-side unique numaralandirma stratejilerinden birini sec
- [x] Uretilen invoice numarasinin mevcut format beklentileri ile uyumlu kalmasini sagla veya degisikligi belgeye isle

### Gorev 4.2: Eszamanli invoice uretimi senaryolarini test et
- [ ] Ayni anda birden fazla invoice olusturuldugunda unique constraint ihlali olusmadigini test et
- [ ] Yuk altinda invoice number uretim maliyetini once/sonra karsilastir

**Kabul Kriterleri**
- [ ] Invoice number uretimi tablo buyuklugune lineer bagli degil
- [ ] Eszamanli olusumda duplicate invoice number olusmuyor

---

## 5. [YUKSEK] Read Path Sorgu ve Mapping Optimizasyonlari

### Gorev 5.1: Read-only sorgularda `AsNoTracking` kullan
- [x] `PlanRepository.GetAllActiveAsync` icin `AsNoTracking` ekle
- [x] `SubscriptionRepository.GetByIdAsync` ve `GetActiveByUserIdAsync` icin kullanim senaryolarina gore read-only varyantlar ekle
- [x] `InvoiceRepository.GetBySubscriptionIdAsync` ve `PaymentRepository.GetByUserIdAsync` gibi endpoint odakli sorgularda tracking ihtiyacini kaldir

### Gorev 5.2: Entity yukleyip sonra map etme yerine projection kullan
- [x] `GetPlans`, `GetMySubscription`, `GetInvoices`, `GetPaymentHistory` sorgularini DTO projection ile optimize et
- [x] AutoMapper `ProjectTo` veya dogrudan `Select` projection yaklasimlarindan repo stiline en uygun olani sec
- [x] Kolon bazli veri cekisini azaltarak sorgu payload'ini kucult

### Gorev 5.3: Payment history sorgusundaki gereksiz `Include` maliyetini kaldir
- [x] `PaymentRepository.GetByUserIdAsync` icindeki `Include(p => p.Subscription)` ihtiyacini yeniden degerlendir
- [x] Sadece filtre icin gereken join ile yetin; navigation entity materialization yapma
- [ ] Pagination altinda sorgu planini ve siralama maliyetini olc

**Kabul Kriterleri**
- [ ] Read endpoint'lerde tracking kapali veya minimum seviyede
- [ ] Payment history ve plans sorgulari daha az kolon ve daha az entity materialization ile calisiyor
- [ ] Endpoint davranisi degismeden response semasi korunuyor

---

## 6. [ORTA] Subscription ve Outbox Erişim Desenlerine Uygun Indeksler

### Gorev 6.1: Subscription sorgulari icin composite index stratejisi ekle
- [x] `GetActiveByUserIdAsync` icin `(user_id, status)` veya uygun partial index stratejisi belirle
- [x] `GetExpiredSubscriptionsAsync` icin `(status, period_end)` veya `status = 'Active'` partial index degerlendir
- [x] Yeni indeksler icin migration ekle ve mevcut veri ile uyumlulugunu dogrula

### Gorev 6.2: Payment ve invoice listeleme sorgulari icin siralama dostu indexleri gozden gecir
- [x] `payments` tablosunda `subscription_id + created_at desc` benzeri composite index ihtiyacini degerlendir
- [x] `invoices` tablosunda `subscription_id + issued_at desc` icin uygun index tanimla
- [ ] Sayfalama performansina etkisini olc

**Kabul Kriterleri**
- [ ] Sorgular tek kolonlu genel indexler yerine erisim desenine daha uygun indexlerden faydalaniyor
- [ ] Listeleme ve worker sorgularinda plan maliyetleri dusuruldu

---

## 7. [ORTA] EF Core Konfigurasyon ve Runtime Iyilestirmeleri

### Gorev 7.1: `AddDbContextPool` uygunlugunu degerlendir ve gerekiyorsa uygula
- [x] `SubscriptionDbContext` icin pooling ile uyumsuz stateful kullanim olup olmadigini kontrol et
- [x] Uygunsa `AddDbContext` yerine `AddDbContextPool` kullan
- [ ] Worker scope'lari ve request tabanli kullanimlarda davranisin degismedigini dogrula

### Gorev 7.2: Gozlemlenebilir performans olcumleri ekle
- [x] Outbox batch suresi, renewal batch suresi ve saga write adim sayisi icin metrik veya structured log alanlari ekle
- [x] Once/sonra performans karsilastirmasi yapilabilecek minimum olcum setini tanimla

**Kabul Kriterleri**
- [ ] DbContext pooling uygulanmissa davranis bozulmadi
- [ ] Yapilan iyilestirmelerin etkisi log/metric ile izlenebilir halde

---

## 8. Onerilen Uygulama Sirasi

### Faz A - Hemen ele alinacaklar
- [x] Gorev 1.1
- [x] Gorev 1.2
- [x] Gorev 2.1
- [x] Gorev 3.1
- [x] Gorev 3.2

### Faz B - Yuksek etkili devam calismalari
- [x] Gorev 4.1
- [x] Gorev 5.1
- [x] Gorev 5.2
- [x] Gorev 6.1

### Faz C - Sertlestirme ve olcum
- [x] Gorev 1.3
- [x] Gorev 2.2
- [x] Gorev 2.3
- [ ] Gorev 4.2
- [x] Gorev 6.2
- [x] Gorev 7.1
- [x] Gorev 7.2
