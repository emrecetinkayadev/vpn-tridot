# hCaptcha ve Rate Limit Prod Konfigürasyonu

Bu runbook, üretim ortamında hCaptcha anahtarlarının temin edilmesi ve API rate limit kurallarının uygulanması adımlarını kapsar.

## 1. hCaptcha Anahtarları

1. **Kurumsal hesap açın**: `https://dashboard.hcaptcha.com` üzerinde prod projesi oluşturun.
2. **Site key / Secret key**: Yeni site yaratırken `sitekey` ve `secret` değerlerini alın.
3. **Ortam ayrımı**: `staging` ve `prod` için ayrı site key oluşturun; rate limitten bağımsız olarak score eşikleri değişebilir.
4. **Vault/SOPS’e yazın**:
   ```bash
   vault kv put kv/vpn-backend/prod \
     HCAPTCHA_ENABLED=true \
     HCAPTCHA_SECRET=<hcaptcha-secret> \
     HCAPTCHA_SITEKEY=<hcaptcha-sitekey> \
     HCAPTCHA_SCORE_THRESHOLD=0.7
   ```
   Staging için ayrı path kullanın (`kv/vpn-backend/staging`). Score eşiklerini sahadaki false-positive oranına göre ayarlayın.
5. **Frontend env**: `NEXT_PUBLIC_HCAPTCHA_SITEKEY` değerini frontend deployment pipeline’ında set edin ki login/signup bileşenleri doğru key’i kullanabilsin.

## 2. Rate Limit Kuralları

Backend üç ana grubu destekliyor:

| Prefix | Amaç | Varsayılan |
|--------|------|------------|
| `RATE_LIMIT_AUTH` | Login, signup, captcha doğrulama | 10 RPS / burst 20 |
| `RATE_LIMIT_CHECKOUT` | Stripe/Iyzico checkout uçları | 5 RPS / burst 10 |
| `RATE_LIMIT_PEERS` | Peer CRUD işlemleri | 8 RPS / burst 16 |

### 2.1 Global RPS

Tüm servis için ana sınırlar:
```env
RATE_LIMIT_RPS=25
RATE_LIMIT_BURST=60
```

### 2.2 Prod Güncellemesi

Vault üzerinde güncellenecek örnek:
```bash
vault kv patch kv/vpn-backend/prod \
  RATE_LIMIT_RPS=25 \
  RATE_LIMIT_BURST=60 \
  RATE_LIMIT_AUTH_ENABLED=true \
  RATE_LIMIT_AUTH_RPS=8 \
  RATE_LIMIT_AUTH_BURST=16 \
  RATE_LIMIT_CHECKOUT_ENABLED=true \
  RATE_LIMIT_CHECKOUT_RPS=3 \
  RATE_LIMIT_CHECKOUT_BURST=6 \
  RATE_LIMIT_PEERS_ENABLED=true \
  RATE_LIMIT_PEERS_RPS=6 \
  RATE_LIMIT_PEERS_BURST=12
```

### 2.3 İzleme

- Prometheus grafikleri: `vpn_backend_http_requests_total{status}` ve `http_request_duration_seconds` metriklerini izleyin.
- Rate limit tetiklenince `429` durum kodları artar; `BackendErrorRateHigh` uyarısı 5xx odaklıdır ancak 429 artışları Grafana panosuna eklenmeli.
- Loki’de `rate limit exceeded` log satırlarını arayarak saldırı/güvenlik bulgusu çıkarın.

## 3. Deploy Sonrası Kontrol Listesi

1. Backend podlarını rolling şekilde yeniden başlatın.
2. Login/signup akışında hCaptcha widget’ının doğru site key ile çalıştığını doğrulayın.
3. Stripe/Iyzico checkout akışını tetikleyip captcha doğrulamasının etkilenmediğini gözlemleyin.
4. Grafana’da 429 trendi beklenen seviyede mi kontrol edin.

## 4. Sorun Giderme

- hCaptcha doğrulaması başarısızsa: 
  - Dashboard’dan site key’in domain kısıtlamalarını kontrol edin.
  - Backend loglarında `hcaptcha verification failed` mesajı var mı bakın.
- Aşırı rate limit: global limitleri geçici olarak yükseltmek için Vault’taki `RATE_LIMIT_*` değerlerini güncelleyin ve redeploy yapın.

## 5. Bağlantılı Dosyalar

- `backend/internal/config/config.go`
- `backend/internal/server/middleware/ratelimit.go`
- `docs/SECRETS.md`
- `deploy/.env.production.example`
