# mTLS CA ve Sertifika Zinciri Runbook

Bu runbook, TriDot VPN kontrol düzlemi ve node agent iletişimi için mutual TLS sertifikalarının nasıl üretileceğini ve rotasyonunu açıklar. Staging ve prod ortamları için ayrı kök sertifikaları kullanmak zorunludur.

## 1. Genel İlkeler

- **Ayrı CA**: `staging` ve `prod` için ayrı kök CA kullanın, anahtarları paylaşmayın.
- **Süreler**: CA sertifikaları 24 ay, sunucu/istemci sertifikaları en fazla 12 ay geçerlilikte olmalı.
- **Saklama**: CA özel anahtarlarını yalnızca Vault/Secret Manager üzerinde saklayın; repo veya CI loglarında bulundurmayın.
- **İzinler**: Backend yalnızca server sertifikasını ve CA zincirini görmeli; node agent yalnızca istemci sertifikasını.

## 2. Sertifika Üretimi

Varsayılan script: `scripts/mtls/generate.sh`

```bash
# staging ortamı için CA + server/agent sertifikaları
./scripts/mtls/generate.sh staging api.staging.tridot.dev agent.staging.tridot.dev

# prod için
./scripts/mtls/generate.sh prod api.tridot.dev agent.tridot.dev
```

Script çıktısı `scripts/mtls/build/<env>/` altında şu dosyaları üretir:

- `ca.crt.pem`, `ca.key.pem`
- `server.crt.pem`, `server.key.pem`
- `client.crt.pem`, `client.key.pem`

> Not: Script yalnızca eksik dosyaları üretir; rotasyon sırasında dizini temizleyerek yeniden üretin.

### 2.1 Vault/SOPS’e Yükleme

- Vault:
  ```bash
  vault kv put kv/vpn-backend/staging MTLS_CA_PEM=@ca.crt.pem MTLS_SERVER_CERT=@server.crt.pem MTLS_SERVER_KEY=@server.key.pem
  vault kv put kv/vpn-node-agent/staging MTLS_CA_PEM=@ca.crt.pem MTLS_CLIENT_CERT=@client.crt.pem MTLS_CLIENT_KEY=@client.key.pem
  ```
- SOPS (YAML örneği):
  ```yaml
  mTLS:
    caPem: |
      -----BEGIN CERTIFICATE-----
      ...
  ```

## 3. Dağıtım Adımları

1. **Hazırlık**
   - Yeni sertifikaları üretin ve gizli depolama alanına yükleyin.
   - CI/CD pipeline’da staging için secret mountlarını güncelleyin.

2. **Staging Doğrulaması**
   - Backend ve node agent podlarını sırayla yeniden başlatın.
   - Backend loglarında `tls: handshake` hatası olup olmadığını kontrol edin.
   - Prometheus’ta `node_agent_wireguard_handshake_ratio` ve `vpn_backend_http_requests_total{status="200"}` metriklerini izleyin.

3. **Prod Rotasyonu**
   - Traffiği canary yaklaşımıyla %10’luk podlara yönlendirin.
   - Eski CA’yı backend konfigürasyonunda ikincil CA olarak tutun (örn. `MTLS_ADDITIONAL_CA` env değeri) ve yeni CA ile handshake’in başarılı olduğunu doğrulayın.
   - Tüm podlar yeni sertifikalarla çalıştıktan sonra eski CA’yı kaldırın.

## 4. Rollback Stratejisi

- Yeni sertifikalar başarısız olursa Vault/SOPS üzerinde eski CA ve sertifikaları geri yükleyin.
- Podları tekrar başlatın ve loki loglarında hata düzeldi mi kontrol edin.
- Post-mortem hazırlayın ve script/config güncellemelerini gözden geçirin.

## 5. Otomasyon ve Güvenlik Notları

- GitHub Actions pipeline’ına sertifika süresi yaklaştığında uyarı verecek cron job ekleyin (`cert-expiry` aracıyla).
- `ops/prometheus/rules/node-health.yml` uyarılarına ek olarak, cert expiry metrikleri için Prometheus exporter düşünün.
- Scriptte üretilecek anahtarlar dosya sisteminde kısa süre kalacağından işlem sonunda dizini silebilirsiniz:
  ```bash
  rm -rf scripts/mtls/build/staging
  ```

## 6. Bağlantılı Dokümanlar

- `docs/SECRETS.md`
- `ops/prometheus/prometheus.yml`
- `node-agent/internal/transport/mtls.go`

## 7. SSS

- **S:** CA(Key) repo’ ya alınabilir mi?
  **C:** Hayır, output klasörü `.gitignore` altında olmalı. Vault/SOPS harici bir yerde saklamayın.
- **S:** Prod rotasyonu sırasında kısa kesinti olur mu?
  **C:** Eski CA’yı ek doğrulama listesinde tutarak kesintisiz geçiş mümkündür.
