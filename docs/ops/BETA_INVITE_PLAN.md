# 100 Beta Kullanıcı Davet Planı

## Amaç
İlk MVP sürümünde 100 beta kullanıcısını davet ederek aktivasyon metriklerini ölçmek.

## Segmentler
- 40 → Mevcut referans müşteriler (gizlilik meraklıları)
- 30 → Start-up CTO toplulukları
- 30 → Güvenlik profesyonelleri (Slack/Discord toplulukları)

## Süreç
1. Airtable veya Notion CRM tablosu hazırlayın (`email`, `kaynak`, `durum`).
2. Davet mailini Postmark ile gönderin (şablon: `ONBOARDING_EMAILS.md`).
3. Aktivasyon adımlarını ölçmek için backend event log’larını (signup, checkout, peer-create) Segment/Amplitude’a iletin.
4. Haftalık olarak `activated users / invited users` metriğini raporlayın.

## Ölçüm
- Aktivasyon kriteri: `signup → cihaz oluşturma` < 2 gün.
- Davet başına dönüşüm oranı %40 hedefleniyor.

## Sorumluluk
- Growth Lead: Davet listesi
- Ops: Aktivasyon metriğini analiz etme
