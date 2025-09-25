# JWT Anahtarı Rotasyonu Runbook

Bu runbook, TriDot VPN backend uygulamasında kullanılan `JWT_SECRET` değerinin güvenli şekilde rotasyonunu açıklar. Amaç, kimlik doğrulama tokenlarının imzasını sağlayan anahtarı düzenli aralıklarla ve olay müdahalelerinde riske atmadan değiştirmektir.

## Önkoşullar

- Vault veya SOPS üzerinden saklanan mevcut `JWT_SECRET` değeri.
- Yeni anahtar üretimi için `openssl` veya `age-keygen`/`pwgen` benzeri bir araç.
- Backend pod/servislerinin health check ve hata loglarını izlemek için erişim (Grafana, Loki, kubectl, vb.).
- Deployment ortamı için gerekli `POSTGRES_DSN`, `REDIS_URL` gibi gizli değerler elde hazır.

## Rotasyon Mimiği

1. **Anahtar üretimi**
   - 32+ baytlık rastgele bir dize üretin:
     ```bash
     openssl rand -base64 48 | tr -d '\n'
     ```
   - Bu değeri `jwt-secret-YYYYMMDD` adıyla Vault/SOPS kaynağına yazın.

2. **Çift Anahtarlı Geçiş (opsiyonel ama önerilen)**
   - Backend yapılandırması `JWT_SECRET_SECONDARY` gibi bir ikinci anahtar destekliyorsa:
     - Yeni anahtarı `JWT_SECRET_SECONDARY` olarak ekleyin.
     - Uygulamayı yeniden dağıtın. Yeni tokenlar ikincil anahtarla imzalanır, mevcut tokenlar birincil ile doğrulanır.
     - 24 saat sonra mevcut aktif oturumların süresi dolduğunda asıl anahtarı yeni değerle değiştirin ve ikincil alanı temizleyin.
   - İkincil anahtar desteği yoksa bir sonraki adıma geçin ve kullanıcıları kısa süreli yeniden oturum açma zorlamasını kabul edin.

3. **Gizli Değeri Güncelleme**
   - Vault:
     ```bash
     vault kv put kv/vpn-backend/prod JWT_SECRET=<yeni-deger>
     ```
   - SOPS (JSON):
     ```bash
     export SOPS_AGE_KEY_FILE=~/.config/sops/age/keys.txt
     sops updatekeys secrets/backend.prod.json
     sops secrets/backend.prod.json
     # JWT_SECRET alanını yeni değerle değiştirin ve kaydedin
     ```

4. **Dağıtım**
   - `deploy/docker` veya Kubernetes manifestlerini kullanarak backend’i yeniden başlatın.
   - Rolling deploy tercih edin (ör. `kubectl rollout restart deployment/backend-api`).
   - Gerekirse 50/50 canary uygulayın: yeni `JWT_SECRET` ile çalışan podu sınırlı trafiğe açın, hataları izleyin.

5. **Doğrulama**
   - Prometheus `vpn_backend_http_requests_total{status="401"}` metriğinde olağan dışı artış var mı kontrol edin.
   - Loki’de `token signature invalid` veya benzeri hatalar için sorgu çalıştırın.
   - Örnek kullanıcıyla login akışını test edin.

6. **Temizlik**
   - Eğer ikincil anahtar kullandıysanız ve her şey yolundaysa eski anahtarı Vault/SOPS kaynağından silin.
   - Eski anahtarı loglar veya dokümanlarda bulundurmayın.

## Olağanüstü Durum / Rollback

- Yeni anahtar nedeniyle toplu 401/403 hataları alınıyorsa eski anahtarı Vault/SOPS üzerinden geri yazın ve redeploy yapın.
- Ağ ekipleri ile koordinasyon halinde kullanıcı bilgilendirmesi yayınlayın.
- Sorun giderildikten sonra yukarıdaki adımları tekrar izleyerek rotasyonu doğru şekilde tamamlayın.

## Otomasyon Notları

- GitHub Actions `release` pipeline’ına isteğe bağlı bir job eklenerek anahtar rotasyonu sonrası otomatik sağlık kontrolleri tetiklenebilir.
- Vault Policy ile yalnızca belirli servis hesaplarının `JWT_SECRET` yazma yetkisi olmasını sağlayın.
- Rotasyon sıklığı: minimum 90 gün, güvenlik incident’ı halinde derhal.

## Referanslar

- `docs/SECRETS.md`
- `ops/prometheus/rules/node-health.yml` (WebhookFailureBurst ve BackendErrorRateHigh uyarıları)
- `TODO.md` Güvenlik Sertleştirme bölümü
