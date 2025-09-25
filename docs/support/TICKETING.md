# Ticketing Entegrasyonu — HelpScout

## Mimari
- Uygulama içindeki “Contact Ops” butonu `support@tridot.dev` adresine yönlenir.
- HelpScout mailbox ile bu adres ilişkilendirilir.
- Zapier/Ops webhook ile P1/P2 ticket’lar Slack #ops-alerts kanalına gönderilir.

## Kurulum Adımları
1. HelpScout üzerinden yeni mailbox oluşturun (`support@tridot.dev`).
2. SPF/DKIM kayıtlarını DNS’e ekleyerek teslimat güvenilirliğini artırın.
3. HelpScout’ta `Doc API Key` oluşturun; backend’e `HELPSCOUT_API_KEY` olarak Vault’a ekleyin.
4. Webhook kurun: Mailbox → Settings → Webhooks → yeni webhook (event: `convo.assigned`, URL: `https://ops.tridot.dev/helpdesk-webhook`).
5. Slack entegrasyonu için HelpScout → Apps → Slack → #ops-alerts kanalını bağlayın.

## İş Akışı
- P1/P2: 30 dakika SLA, otomatik Slack bildirimi, on-call ataması.
- P3/P4: 24 saat SLA, Help Center’a yönlendirme.

## İzleme
- Haftalık rapor: çözüm süresi, reopen sayısı.
- Ops metrikleri Grafana “Support Overview” panosuna aktarılacak.
