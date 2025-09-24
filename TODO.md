# TODO — VPN MVP Yol Haritası (6–8 Hafta)

> Biçim: GitHub task list. `[ ]` açık, `[x]` tamam. Alt görevler iç içe listelenmiştir.

## 0) Proje Hazırlık ve Altyapı
- [x] Monorepo oluşturma
  - [x] `backend/`, `frontend/`, `node-agent/`, `infra/`, `deploy/`, `docs/`, `ops/`, `scripts/`
  - [x] CODEOWNERS, `.editorconfig`, `.gitignore`, `.gitattributes`
  - [x] Pre-commit hook’ları (gitleaks, golangci-lint, eslint)
- [ ] CI iskeleti
  - [x] `ci-backend.yaml` build+test
  - [x] `ci-frontend.yaml` build+test+playwright (staging)
  - [x] `ci-agent.yaml` build+unit
  - [x] `security-scan.yaml` (gitleaks, syft/grype)
  - [x] `release.yaml` (tag → image push → GH Release)
- [x] Devcontainer ve VSCode önerileri (`.devcontainer/`, `.vscode/`)

## 1) Backend (Control Plane, Go)
- [x] Proje iskeleti ve temel bağımlılıklar
  - [x] HTTP server (Gin/Fiber), middleware (auth, ratelimit, logging)
  - [x] Config yükleme (env + defaults)
- [x] Veritabanı ve migrasyonlar (Postgres)
  - [x] Şema: users, plans, subscriptions, payments, regions, nodes, peers, sessions
  - [x] Goose/Atlas ile `0001_init.sql`…
  - [x] DB indexleri ve foreign key’ler
- [x] Kimlik ve oturum
  - [x] Sign‑up, login, email doğrulama
  - [x] JWT access/refresh akışı
  - [x] Şifre sıfırlama
  - [x] 2FA/TOTP (v1.1, opsiyon)
- [x] Planlar ve ödeme (Stripe + Iyzico)
  - [x] Plan CRUD (seed: Aylık/3A/Yıllık)
  - [x] Checkout oturumu oluşturma
  - [x] Webhook doğrulama ve idempotent işlem
  - [x] Abonelik durum geçişleri (`trialing|active|canceled|past_due`)
  - [x] Fatura geçmişi görünümü (özet meta)
- [x] Bölge ve kapasite
  - [x] Regions listesi (TR‑IST, TR‑IZM, EU-FRA/NL)
  - [x] Node kayıt/sağlık uç noktaları
  - [x] Kapasite puanı hesaplama (aktif peer, throughput, CPU)
- [x] Peer/cihaz yönetimi
  - [x] Peer CRUD (kullanıcı başına cihaz limiti)
  - [x] Client‑side public key kabulü ve doğrulama
  - [x] Server‑side key üretimi (opsiyon) + tek seferlik indirme
  - [x] Config üretimi (AllowedIPs, DNS, MTU, keepalive)
  - [x] İmzalı **tek‑kullanımlık** config URL’leri (TTL: 24h)
  - [x] QR kod oluşturma (PNG/SVG)
- [x] Kullanım görünürlüğü
  - [x] Son bağlantı zamanı, toplam up/down bayt
  - [x] Cihaz bazlı silme/yeniden adlandırma
- [x] Observability
  - [x] Prometheus metrics endpoint
  - [x] Yapılandırılabilir request logging + maskleme
- [x] Güvenlik
  - [x] Ratelimit (auth, checkout, peers)
  - [x] hCaptcha entegrasyonu (signup/login) — basit skor doğrulama
  - [x] CORS/CSRF ayarları
  - [x] Secret yönetimi: SOPS/Vault entegrasyonu
- [x] Testler
  - [x] Unit: auth, webhook, peers
  - [x] Integration: config üretimi, tek‑kullanımlık URL
  - [x] E2E: signup→checkout (mock)→peer create→config fetch

## 2) Node Agent (Go)
- [x] İskelet
  - [x] mTLS client, token doğrulama
  - [x] Konfig yükleme (`env`, `file`)
- [x] WireGuard entegrasyonu
  - [x] `wg0` oluşturma, port yönetimi
  - [x] Peer ekleme/çıkarma, kalıcılık (`/etc/wireguard/*.conf`)
  - [x] MTU/keepalive/DNS ayarları
  - [x] iptables NAT + kill‑switch kuralları
- [x] Sağlık ve telemetri
  - [x] Handshake oranı, aktif peer sayısı, NIC throughput
  - [x] Prometheus exporter
- [x] Dayanıklılık
  - [x] Crash‑safe state, retry/backoff
  - [x] Drain modu (yeni peer kabul etme)
- [ ] Testler
  - [ ] Unit: wg sarmalayıcı, health reporter
  - [ ] Entegrasyon: backend mTLS çağrıları (mock CA)

## 3) Frontend (Next.js 15)
- [x] UI iskeleti, tema, navigasyon
- [ ] Auth sayfaları (signup/login, şifre sıfırlama)
  - [x] Login / signup ekran şablonları
  - [ ] Şifre sıfırlama akışı
- [ ] Planlar ve ödeme akışı (Stripe/Iyzico checkout)
  - [x] Plan kartları + checkout CTA placeholder
  - [ ] Stripe/Iyzico checkout entegrasyonu
- [ ] Bölgeler ve kapasite görünümü
  - [x] Bölge tablosu ve kapasite yer tutucuları
  - [ ] Prometheus/agent verileri ile besleme
- [ ] Cihazlar/Peers
  - [x] Cihaz listesi ve görev kuyruğu şablonu
  - [ ] Listele, oluştur, sil, yeniden adlandır
  - [ ] QR/CONF indirme butonları
- [ ] Hesap/kullanım sayfası (toplam trafik, son bağlantı)
  - [x] Profil ve kullanım blokları (placeholder)
  - [ ] Gerçek API entegrasyonu
- [ ] Hata ve durum sayfaları
- [ ] E2E testleri (Playwright): cihaz oluşturma akışı
  - [x] Smoke testi: dashboard başlığının görünmesi

## 4) Altyapı Otomasyonu
- [ ] Terraform
  - [ ] VPC, subnet, IGW, route, security group
  - [ ] VM/metal node’lar + statik IP’ler
  - [ ] DNS kayıtları (A/AAAA)
- [ ] Ansible
  - [ ] Kernel parametreleri (ip_forward, rp_filter)
  - [ ] Paketler: `wireguard-tools`, `iptables`, `chrony`, `docker`
  - [ ] Agent deploy (systemd veya Docker)
  - [ ] Prometheus target ekleme
- [ ] Ortamlar
  - [ ] `staging` cluster/node’lar
  - [ ] `prod` cluster/node’lar

## 5) Ödeme ve Hukuki
- [ ] Stripe ve Iyzico test hesapları
- [ ] Plan/kuvvetli 3D ve BKM onayı
- [ ] Webhook güvenliği (imza doğrulama)
- [ ] KVKK/GDPR metinleri (`docs/privacy.md`)
- [ ] KVKK aydınlatma ve açık rıza UI onayı
- [ ] Abonelik iptal/iade akışları

## 6) İzleme, Günlükler, Uyarılar
- [ ] Prometheus + Grafana deploy
  - [ ] Dashboard’lar: nodes-overview, peers-overview
- [ ] Loki + Alertmanager
  - [ ] Kurallar: node down, handshake error spike, webhook failure, error rate
- [ ] Durum sayfası (status page) — basit static

## 7) Güvenlik Sertleştirme
- [ ] mTLS CA ve sertifika zinciri (staging/prod ayrı)
- [ ] JWT anahtarı rotasyonu planı
- [ ] Ratelimit ve hCaptcha prod anahtarları
- [ ] Admin panel IP allowlist
- [ ] SSH hardening, bastion, JIT erişim
- [ ] SBOM üretimi ve imaj taraması (syft/grype)
- [ ] Container imzalama (cosign) — v1.1

## 8) Sızıntı ve Yük Testleri
- [ ] DNS leak testi (IPv4/IPv6)
- [ ] IPv6 leak testi (OSX/iOS/Android)
- [ ] MTU/fragmentation testleri
- [ ] Throughput testi (`iperf3`) node başına hedef 2+ Gbps aggregate
- [ ] Port engelleme senaryoları (UDP varyasyonları)

## 9) Destek ve Operasyon
- [ ] Help Center: SSS, kurulum rehberi (iOS/Android/macOS/Windows/Linux)
- [ ] Biletleme entegrasyonu (HelpScout/Freshdesk)
- [ ] Faturalandırma SSS ve iade politikası
- [ ] Runbook’lar
  - [ ] `incident-node-down.md`
  - [ ] `rotate-wg-keys.md`
  - [ ] `restore-db.md`
  - [ ] `postmortems` şablonu

## 10) Pazarlama Hazırlığı (MVP için asgari)
- [ ] Landing page (planlar, gizlilik, SSS, durum sayfası linki)
- [ ] Onboarding e‑posta şablonları (Postmark/SES)
- [ ] Deneme → ücretli dönüşüm e‑posta akışı
- [ ] Basit marka varlıkları (logo, renkler)

## 11) MVP Çıkış Kontrol Listesi
- [ ] 100 beta kullanıcı için davet listesi ve ölçüm planı
- [ ] Aktivasyon süresi ortalama < 2 dk (ödeme→bağlanma)
- [ ] İlk hafta bağlantı başarısızlık oranı < %2
- [ ] DNS sızıntı testi laboratuvar %100, saha > %98
- [ ] İzleme panoları ve alarmlar canlı
- [ ] KVKK/GDPR metinleri yayında
- [ ] Durum sayfası yayında

## 12) Post‑MVP (v1.1–v1.3)
- [ ] Mobil uygulamalar (iOS/Android native)
- [ ] Tarayıcı eklentisi
- [ ] DNS filtre profilleri (ads/malware)
- [ ] Obfuscation protokolleri (Hysteria2/TUIC/REALITY)
- [ ] Dedicated IP ve port‑forward

---

### Notlar
- Her görev için **kabul kriteri** ve **sahip** ekleyin. PR açılırken görev ID’si ile referans verin.
- Güvenlikle ilgili görevler tamamlanmadan prod’a dağıtım yapılmaz.
