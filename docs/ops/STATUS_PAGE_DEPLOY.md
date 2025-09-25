# Durum Sayfası Yayın Planı

## Frontend
- Next.js route: `frontend/app/(status)/status/page.tsx`
- Veri kaynağı: `fetchStatusSnapshot()` (Prometheus → fallback)

## Dağıtım
1. `NEXT_PUBLIC_STATUS_PROMETHEUS_URL` env değeri production Prometheus endpoint’ine ayarlanır.
2. Vercel veya Netlify’ya static deploy (ISR gerekmiyor).
3. Domain: `status.tridot.dev` (CNAME → hosting provider).

## Sağlık
- Sayfa 1 dakikada bir Prometheus’a istek atar (`no-store`).
- Hata durumunda fallback veri + uyarı banner’ı.

## Kontrol Listesi
- [ ] SSL sertifikası aktif
- [x] Sayfa repo içinde
- [ ] UptimeRobot monitörü eklendi
