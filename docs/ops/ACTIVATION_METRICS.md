# Aktivasyon Süresi Ölçüm Planı (< 2 dk)

## Tanım
Aktivasyon süresi = `checkout_completed_at` - `vpn_connected_at_first_time`.

## Veri Kaynakları
- Backend events: `billing.checkout.completed`, `peers.device.connected`
- Segment event stream veya PostgreSQL audit tablosu (`events`)

## Ölçüm Pipeline
1. Checkout webhook → `events` tablosuna satır ekler.
2. Node agent health raporu ilk handshake’i raporlar → backend `peers.sessions` tablosu güncellenir.
3. Looker/Metabase dashboard:
   ```sql
   SELECT AVG(connected_at - checkout_at) FROM activation_sessions WHERE checkout_at >= now() - interval '7 days';
   ```

## Sınır
- Ortalama < 120 saniye, p95 < 180 saniye.
- Alarm: 3 gün üst üste hedef aşılıyorsa Slack #ops-alerts.
