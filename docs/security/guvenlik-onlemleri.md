# Güvenlik Önlemleri Kontrol Listesi

Prod ortamına çıkmadan önce tamamlanması gereken başlıca güvenlik işleri aşağıda öncelik etiketleriyle listelendi. Her madde ilgili kod/dosya referansını içerir.

- **[Kritik] Node ⇄ API mTLS’i sunucu tarafında zorunlu kılın** — `backend/internal/server/handlers/nodes/handler.go:17` ve HTTP sunucusu genel TLS terminasyonuna bırakılmış durumda. Load balancer / ingress seviyesinde client certificate doğrulaması tanımlayın, başarısız denemeleri kapatın.
- **[Kritik] Provision token’ı döngüsel hale getirin ve IP tabanlı erişim kısıtlayın** — statik `NODE_PROVISION_TOKEN` kullanımı (bkz. `backend/internal/config/config.go:315`, `node-agent/internal/config/config.go:90`). Token’ı Vault üzerinden kısa ömürlü üretin, sadece bastion kaynaklı IP bloklarına izin verin.
- **[Kritik] Node health/register uçları için oran kontrolü + anomali alarmı ekleyin** — şu an hiçbir rate limit / imza kontrolü yok (`backend/internal/server/setup/setup.go:94`). Token sızarsa kapasite skorları oynanabilir.
- **[Kritik] Backend `/metrics` uç noktasını ağ seviyesinde koruyun** — `backend/internal/server/server.go:38` global olarak açıyor. Sadece internal ağdan erişim, Basic Auth veya ayrı port gereklidir.
- **[Kritik] Next.js demo API’lerini üretimde kapatın** — `frontend/app/api/peers/route.ts:10` bellekte sahte peer yaratıyor, auth yok. Build esnasında dev-only bayrakla devre dışı bırakın.
- **[Yüksek] Frontend’e güvenlik başlıkları ekleyin** — `frontend/next.config.ts:3` sadece `reactStrictMode` tanımlı. `Content-Security-Policy`, `Strict-Transport-Security`, `X-Frame-Options`, `X-Content-Type-Options` başlıklarını middleware ile ekleyin.
- **[Yüksek] Agent config dosyasını 0600 izinle tutun** — Ansible şablonu `/etc/tridot/agent.yaml` içinde mTLS private key barındırıyor (`infra/ansible/roles/agent/tasks/main.yml:13`). Dosya izinlerini `0600`, dizini `0700` yapın.
- **[Yüksek] Agent binary dağıtımını imzalı artefaktan yapın** — `infra/ansible/roles/agent/tasks/main.yml:7` doğrudan repo içi ikili kopyalanıyor. Hash doğrulaması + imzalı release kullanın.
- **[Yüksek] Vault TLS doğrulamasını zorunlu kılın** — `backend/internal/platform/secrets/manager.go:27` `InsecureSkipVerify` opsiyonu var. Prod ortamında `VAULT_TLS_SKIP_VERIFY=false` policy’sine uyan bir denetim ekleyin.
- **[Yüksek] Admin IP allowlist’i devreye alın** — `backend/internal/server/middleware/auth.go:38` boş allowlist tüm yönetici uçlarını açıyor. Prod değerlerini config’e ekleyin, Prometheus uyarısı bağlayın.
- **[Orta] Node health verilerini imzalı hale getirin** — Payload’a zaman damgalı HMAC ekleyerek replay/injection riskini azaltın (aynı dosya `nodes/handler.go`).
- **[Orta] Terraform apply öncesi DNS kayıtları için drift denetimi ekleyin** — yeni Route53 modülü (`infra/terraform/modules/dns/main.tf:31`) overwrite yetkisi var; `allow_overwrite=false` ile yanlışlıkla kayıt ezilmesini engelleyin.
- **[Orta] KVKK/GDPR metni canlıya taşınıp panel linki doğrulanmalı** — `docs/privacy.md` ve `frontend/app/(auth)/signup/page.tsx:46` linki gerçek ortamda 200 döndürmeli; monitoring ekleyin.
- **[Orta] Stripe/Iyzico test-hesap erişimlerini sınırlayın** — API anahtarlarını sadece CI’de kullanılan servislere verin; Vault policy’si ekleyin.
- **[Orta] İstemci konfig URL token’ları için IP/UA şablonu takip edin** — Tek kullanımlık token’lar (`backend/internal/peers/service.go:143`) için aşırı indirme denemelerini izleyin, eşik koyun.
- **[Düşük] mtls/generate.sh çıktısını otomatik temizleyin** — `scripts/mtls/generate.sh` build dizininde anahtarları bırakıyor; script çıkışında `rm -rf` ekleyin.

Öncelik etiketleri: **Kritik** (prod öncesi bloklayıcı), **Yüksek** (ilk sprintte), **Orta** (kısa vadeli), **Düşük** (fırsat oldukça).
