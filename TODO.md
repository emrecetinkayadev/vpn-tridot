# TODO — VPN MVP Yol Haritası (6–8 Hafta)

> Biçim: GitHub task list. `[ ]` açık, `[x]` tamam. Alt görevler iç içe listelenmiştir.

## 0) Proje Hazırlık ve Altyapı
- [ ] Monorepo oluşturma
  - [ ] `backend/`, `frontend/`, `node-agent/`, `infra/`, `deploy/`, `docs/`, `ops/`, `scripts/`
  - [ ] CODEOWNERS, `.editorconfig`, `.gitignore`, `.gitattributes`
  - [ ] Pre-commit hook’ları (gitleaks, golangci-lint, eslint)
- [ ] CI iskeleti
  - [ ] `ci-backend.yaml` build+test
  - [ ] `ci-frontend.yaml` build+test+playwright (staging)
  - [ ] `ci-agent.yaml` build+unit
  - [ ] `security-scan.yaml` (gitleaks, syft/grype)
  - [ ] `release.yaml` (tag → image push → GH Release)
- [ ] Devcontainer ve VSCode önerileri (`.devcontainer/`, `.vscode/`)

## 1) Backend (Control Plane, Go)
- [ ] Proje iskeleti ve temel bağımlılıklar
  - [ ] HTTP server (Gin/Fiber), middleware (auth, ratelimit, logging)
  - [ ] Config yükleme (env + defaults)
- [ ] Veritabanı ve migrasyonlar (Postgres)
  - [ ] Şema: users, plans, subscriptions, payments, regions, nodes, peers, sessions
  - [ ] Goose/Atlas ile `0001_init.sql`…
  - [ ] DB indexleri ve foreign key’ler
- [ ] Kimlik ve oturum
  - [ ] Sign‑up, login, email doğrulama
  - [ ] JWT access/refresh akışı
  - [ ] Şifre sıfırlama
  - [ ] 2FA/TOTP (v1.1, opsiyon)
- [ ] Planlar ve ödeme (Stripe + Iyzico)
  - [ ] Plan CRUD (seed: Aylık/3A/Yıllık)
  - [ ] Checkout oturumu oluşturma
  - [ ] Webhook doğrulama ve idempotent işlem
  - [ ] Abonelik durum geçişleri (`trialing|active|canceled|past_due`)
  - [ ] Fatura geçmişi görünümü (özet meta)
- [ ] Bölge ve kapasite
  - [ ] Regions listesi (TR‑IST, TR‑IZM, EU‑FRA/NL)
  - [ ] Node kayıt/sağlık uç noktaları
  - [ ] Kapasite puanı hesaplama (aktif peer, throughput, CPU)
- [ ] Peer/cihaz yönetimi
  - [ ] Peer CRUD (kullanıcı başına cihaz limiti)
  - [ ] Client‑side public key kabulü ve doğrulama
  - [ ] Server‑side key üretimi (opsiyon) + tek seferlik indirme
  - [ ] Config üretimi (AllowedIPs, DNS, MTU, keepalive)
  - [ ] İmzalı **tek‑kullanımlık** config URL’leri (TTL: 24h)
  - [ ] QR kod oluşturma (PNG/SVG)
- [ ] Kullanım görünürlüğü
  - [ ] Son bağlantı zamanı, toplam up/down bayt
  - [ ] Cihaz bazlı silme/yeniden adlandırma
- [ ] Observability
  - [ ] Prometheus metrics endpoint
  - [ ] Yapılandırılabilir request logging + maskleme
- [ ] Güvenlik
  - [ ] Ratelimit (auth, checkout, peers)
  - [ ] hCaptcha entegrasyonu (signup/login) — basit skor doğrulama
  - [ ] CORS/CSRF ayarları
  - [ ] Secret yönetimi: SOPS/Vault entegrasyonu
- [ ] Testler
  - [ ] Unit: auth, webhook, peers
  - [ ] Integration: config üretimi, tek‑kullanımlık URL
  - [ ] E2E: signup→checkout (mock)→peer create→config fetch

## 2) Node Agent (Go)
- [ ] İskelet
  - [ ] mTLS client, token doğrulama
  - [ ] Konfig yükleme (`env`, `file`)
- [ ] WireGuard entegrasyonu
  - [ ] `wg0` oluşturma, port yönetimi
  - [ ] Peer ekleme/çıkarma, kalıcılık (`/etc/wireguard/*.conf`)
  - [ ] MTU/keepalive/DNS ayarları
  - [ ] iptables NAT + kill‑switch kuralları
- [ ] Sağlık ve telemetri
  - [ ] Handshake oranı, aktif peer sayısı, NIC throughput
  - [ ] Prometheus exporter
- [ ] Dayanıklılık
  - [ ] Crash‑safe state, retry/backoff
  - [ ] Drain modu (yeni peer kabul etme)
- [ ] Testler
  - [ ] Unit: wg sarmalayıcı, health reporter
  - [ ] Entegrasyon: backend mTLS çağrıları (mock CA)

## 3) Frontend (Next.js 15)
- [ ] UI iskeleti, tema, navigasyon
- [ ] Auth sayfaları (signup/login, şifre sıfırlama)
- [ ] Planlar ve ödeme akışı (Stripe/Iyzico checkout)
- [ ] Bölgeler ve kapasite görünümü
- [ ] Cihazlar/Peers
  - [ ] Listele, oluştur, sil, yeniden adlandır
  - [ ] QR/CONF indirme butonları
- [ ] Hesap/kullanım sayfası (toplam trafik, son bağlantı)
- [ ] Hata ve durum sayfaları
- [ ] E2E testleri (Playwright): cihaz oluşturma akışı

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