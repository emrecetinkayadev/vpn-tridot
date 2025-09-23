# VPN MVP — README

Bu depo, PRD’de tanımlanan tüketici VPN servisinin **MVP** sürümü için monorepo yapısını, kurulum adımlarını, çalışma talimatlarını, güvenlik ve operasyon rehberini içerir.

> Hedef: 6–8 haftada üretime hazır, otomatik sağlanan WireGuard tabanlı bir VPN servisi.

---

## İçindekiler

* [Mimari Genel Bakış](#mimari-genel-bakış)
* [Özellikler](#özellikler)
* [Depo Yapısı](#depo-yapısı)
* [Önkoşullar](#önkoşullar)
* [Ortam Değişkenleri](#ortam-değişkenleri)
* [Hızlı Başlangıç](#hızlı-başlangıç)
* [Backend API](#backend-api)
* [Node Agent](#node-agent)
* [Frontend](#frontend)
* [Veritabanı ve Migrasyonlar](#veritabanı-ve-migrasyonlar)
* [Ödeme Entegrasyonları](#ödeme-entegrasyonları)
* [İzleme ve Günlükler](#izleme-ve-günlükler)
* [Güvenlik](#güvenlik)
* [KVKK/GDPR ve Gizlilik](#kvkkgdpr-ve-gizlilik)
* [Test Stratejisi](#test-stratejisi)
* [Yayınlama ve DevOps](#yayınlama-ve-devops)
* [Altyapı Otomasyonu](#altyapı-otomasyonu)
* [Sızıntı ve Yük Testleri](#sızıntı-ve-yük-testleri)
* [Kapasite Planlama](#kapasite-planlama)
* [Sorun Giderme](#sorun-giderme)
* [Yol Haritası](#yol-haritası)
* [Sık Sorulanlar](#sık-sorulanlar)
* [Lisans](#lisans)

---

## Mimari Genel Bakış

**Bileşenler**

* **Control Plane (Backend API, Go):** Kimlik, abonelik, ödeme, peer yönetimi, kapasite ve konfig üretimi. REST dışa, gRPC içe.
* **Node Agent (Go):** Her VPN node’unda çalışan hafif daemon. wgctrl ile peer ekleme/çıkarma, health check, ölçüm gönderimi.
* **Data Plane (WireGuard):** UDP tabanlı tünel trafiği. Kernel mod.
* **Frontend (Next.js):** Kullanıcı paneli, planlar, config/QR üretimi ve indirme.
* **İzleme:** Prometheus, Grafana, Loki, Alertmanager.
* **İnfra:** Terraform (VPC, subnet, firewall, VM), Ansible (kernel, wireguard-tools, agent deploy).

```
Kullanıcı ↔ Frontend (Next.js) ↔ Backend (Go, REST/gRPC) ↔ Postgres/Redis
                                            ↘
                                             Node Agent (Go, mTLS) ↔ WireGuard
```

---

## Özellikler

* Tek tıkla cihaz oluşturma ve **QR/CONF** indirme
* 5 cihaza kadar eşzamanlı kullanım
* Bölge seçimi ve **kapasite puanı** önerisi
* Trafik ve içerik logu yok; sadece oturum toplamları
* Stripe + Iyzico ile abonelik
* Terraform + Ansible ile tam otomatik node sağlama
* Prometheus ve Grafana panoları

---

## Depo Yapısı

```
repo-root/
  frontend/
    app/
    components/
    lib/
    public/
    package.json
  backend/
    cmd/api/
    internal/
      auth/         # JWT, 2FA (v1.1)
      billing/      # Stripe/Iyzico, webhook işleyicileri
      peers/        # WG peer yaşam döngüsü
      regions/      # Bölge/kapasite yönetimi
      nodes/        # Node kayıt/sağlık
      storage/      # Postgres, Redis
      wg/           # wgctrl sarmalayıcıları
    proto/
    go.mod
  node-agent/
    cmd/agent/
    internal/
      health/
      wg/
      rpc/
    go.mod
  infra/
    terraform/
      modules/
      envs/
    ansible/
      roles/
      playbooks/
  deploy/
    docker/
      backend.Dockerfile
      agent.Dockerfile
      compose.node.yaml
    k8s/
  scripts/
    generate_wg_conf.sh
    leaktest.sh
  docs/
    architecture.md
    privacy.md
  .github/workflows/
    ci-backend.yaml
    ci-frontend.yaml
    deploy-backend.yaml
```

---

## Önkoşullar

* **Go** 1.22+
* **Node.js** 22+, **pnpm** veya **npm**
* **Docker** 24+
* **PostgreSQL** 16+, **Redis** 7+
* **Terraform** 1.8+, **Ansible** 2.16+
* Bulut erişimi (Oracle/Hetzner vb.)

---

## Ortam Değişkenleri

### Backend

```
# Ayrıntılı örnek dosyalar için `backend/.env.example`, `backend/.env.development` ve `deploy/.env.production.example` dosyalarına bakın.
```

Öne çıkan ayarlar:

- `LOG_LEVEL=info` ve `LOG_REQUESTS_ENABLED=true` isteklere ait loglamayı kontrol eder.
- `LOG_REQUEST_HEADERS`, `LOG_MASK_HEADERS`, `LOG_REQUEST_QUERY_PARAMS`, `LOG_MASK_QUERY_PARAMS` loglarda görünen alanları tanımlar.
- `METRICS_ENABLED=true`, `METRICS_PATH=/metrics`, `METRICS_NAMESPACE=vpn_backend`, `METRICS_SUBSYSTEM=http` Prometheus endpoint yapılandırmasıdır.
- `CORS_*` değişkenleri (origin/method/header listeleri) cross-site isteklerini yönetir; `CSRF_*` ayarları Origin denetimini açar.
- `RATE_LIMIT_*` değerleri (global ve `AUTH/CHECKOUT/PEERS` için `_RPS`/`_BURST`) istek hızını sınırlar.
- `HCAPTCHA_*` anahtarları (secret/sitekey/threshold) signup ve login hCaptcha doğrulamasını yönetir.
- Secrets yönetimi: `SOPS_SECRETS_*` veya `VAULT_*` değişkenleri ile environment değerleri otomatik yüklenir (detaylar `docs/SECRETS.md`).

### Agent

```
CONTROL_PLANE_URL=https://cp.example.com
AGENT_TOKEN=...
AGENT_METRICS_ADDR=:9102
AGENT_STATE_DIR=/var/lib/vpn-agent
AGENT_MAX_RETRY_INTERVAL=2m
MTLS_CA_PEM=base64:...
MTLS_CLIENT_CERT=base64:...
MTLS_CLIENT_KEY=base64:...
WG_INTERFACE=wg0
WG_PORT=51820
WG_ADDRESS=10.0.0.2/32
WG_DNS=1.1.1.1,8.8.8.8
WG_ENABLE_NAT=true
WG_ENABLE_KILLSWITCH=false
```

### Frontend

```
NEXT_PUBLIC_API_BASE=https://api.example.com
NEXT_PUBLIC_STRIPE_PK=pk_test_...
```

> Not: Gizli bilgiler için SOPS/age veya Vault kullanın. `.env` dosyalarını commit etmeyin.

---

## Hızlı Başlangıç

1. Depoyu klonlayın.
2. **Backend** bağımlılıkları ve build:

   ```bash
   cd backend && go mod download && go build ./cmd/api
   ```
3. **Frontend** bağımlılıkları ve dev:

   ```bash
   cd frontend && pnpm i && pnpm dev
   ```
4. **DB** ve **Redis**’i çalıştırın (lokal veya Docker). Migrasyonları uygulayın.
5. **Backend**’i çalıştırın:

   ```bash
   ./api
   ```
6. **Agent**’ı bir test node’unda çalıştırın. WireGuard’ı kurun ve `compose.node.yaml` ile ayağa kaldırın.

---

## Backend API

### Çalıştırma

```bash
cd backend
GOOSE_DRIVER=postgres GOOSE_DBSTRING="$POSTGRES_DSN" goose up   # migrasyon
go run ./cmd/api
```

### Uç Noktalar (Özet)

```
POST   /api/v1/auth/signup
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/password-reset/request
POST   /api/v1/auth/password-reset/confirm
GET    /api/v1/plans
POST   /api/v1/checkout
POST   /api/v1/webhooks/stripe
GET    /api/v1/regions
POST   /api/v1/nodes/register
POST   /api/v1/nodes/health
GET    /api/v1/peers
GET    /api/v1/peers/usage
POST   /api/v1/peers
PATCH  /api/v1/peers/{peerId}
DELETE /api/v1/peers/{peerId}
GET    /api/v1/peers/config/{token}  # tek-kullanımlık imzalı URL + QR
GET    /api/v1/account/payments
```

### Örnek Yanıt: `/api/v1/peers/config/{token}`

```json
{
  "config": "[Interface]..."
}
```

### Örnek Yanıt: `/api/v1/peers/usage`

```json
{
  "total_bytes_tx": 15874432,
  "total_bytes_rx": 20893440,
  "total_bytes": 36767872,
  "peer_count": 3,
  "active_peer_count": 2,
  "last_handshake_at": "2025-01-01T10:15:00Z"
}
```

### Yetkilendirme

* Dış REST istekleri **JWT** taşır.
* Agent↔Backend iletişimi **mTLS** + **token** ile korunur.

---

## Node Agent

### Kurulum

* Hedef node’da: `wireguard-tools`, `iptables`, `docker` hazır olmalı.
* Deploy örneği (`deploy/compose.node.yaml`):

```yaml
services:
  agent:
    image: registry.example.com/vpn/agent:latest
    network_mode: host
    cap_add: ["NET_ADMIN"]
    volumes:
      - /etc/wireguard:/etc/wireguard
      - /var/log/agent:/var/log/agent
    environment:
      - CONTROL_PLANE_URL=${CONTROL_PLANE_URL}
      - AGENT_TOKEN=${AGENT_TOKEN}
```

### Sorumluluklar

* WG arayüz oluşturma (`wg0`), peer ekleme/çıkarma
* Keepalive, MTU, DNS, AllowedIPs konfigleri
* Sağlık metriklerini Prometheus’a export etme
* Arıza durumunda **drain** sinyali
* `AGENT_STATE_DIR` altında peer konfiglerini kalıcı tutma (crash-safe)

### Sağlık ve Metrikler

* Control plane’e gönderilen health payload’ı şu alanları içerir:
  * `peer_count`, `active_peer_count`, `handshake_ratio`
  * `rx_bytes`/`tx_bytes` ve hesaplanan `rx_bps`/`tx_bps`
  * `last_handshake` zaman damgası (UTC)
  * `drain` → node’un yeni peer kabul edip etmediği
* Prometheus endpoint’i `AGENT_METRICS_ADDR` adresinde (`default: :9102`) `/metrics` yolunda aşağıdaki metrikleri sağlar:
  * `node_agent_wireguard_peers`, `node_agent_wireguard_active_peers`
  * `node_agent_wireguard_handshake_ratio`
  * `node_agent_wireguard_rx_bytes_total` / `_tx_bytes_total`
  * `node_agent_wireguard_rx_throughput_bps` / `_tx_throughput_bps`
  * `node_agent_wireguard_last_handshake`
* Ayrıntılı açıklama için `docs/NODE_AGENT_METRICS.md` dosyasına bakın.
* Drain modu için `AGENT_STATE_DIR` altında `drain` dosyasını oluşturmak yeterli (`touch $AGENT_STATE_DIR/drain`); silmek drain’i kapatır.

---

## Frontend

### Teknoloji

* Next.js 15, TypeScript, Tailwind v4, shadcn/ui, TanStack Query

### Komutlar

```bash
pnpm dev      # localhost:3000
pnpm build
pnpm start
```

### Sayfalar

* `/signup`, `/login`
* `/plans` → Stripe checkout (Iyzico devreye alınacak)
* `/regions` → kapasite puanları
* `/devices` → peer listesi, oluştur/sil, QR/CONF indir
* `/account` → kullanım görünümü

---

## Veritabanı ve Migrasyonlar

### Şema (Özet)

* `users(id, email, pass_hash, twofa_secret, created_at)`
* `subscriptions(id, user_id, plan_id, status, renew_at)`
* `plans(id, name, price, device_limit)`
* `payments(id, user_id, provider, external_id, amount, status, created_at)`
* `regions(id, name, country_code)`
* `nodes(id, region_id, public_ip, wg_port, capacity_score, status)`
* `peers(id, user_id, node_id, pubkey, ip4, ip6, name, created_at)`
* `sessions(id, peer_id, bytes_up, bytes_down, last_handshake_at, window)`

### Migrasyon Araçları

* Öneri: `goose` veya `atlas` (SQL)

---

## Ödeme Entegrasyonları

* **Stripe**: Global kartlar, abonelik planları, webhook doğrulama.
* **Iyzico**: İleride devreye alınacak (API anahtarları `.env` içinde yer alır).
* Başarılı ödeme → `subscriptions.status=active` → cihaz/peer hakkı açılır.

---

## İzleme ve Günlükler

* **Prometheus** endpoint: `GET /metrics` (varsayılan), `METRICS_*` ortam değişkenleri ile özelleştirilebilir.
* **HTTP metrikleri**: `http_requests_total`, `http_request_duration_seconds` label seti (method/path/status).
* **İstek logları**: `LOG_REQUEST_*` env ayarları ile maskelenebilir; hassas alanlar `***` olarak yazılır.
* **Grafana** panoları: Bölge ve node bazlı (metrics scrape sonrası).
* **Loki**: API ve Agent günlükleri.
* **Alertmanager**: Node down, webhook failure, error oranı artışı.

Detaylı bilgi için `docs/OBSERVABILITY.md` dosyasına bakın.

---

## Güvenlik

* mTLS, JWT, route bazlı ratelimit, CORS allowlist, SOPS/Vault tabanlı secrets ve hCaptcha doğrulaması
* CSRF guard: origin allowlist’e uymayan POST/PUT/PATCH/DELETE istekleri bloklanır
* Release ikilileri cosign ile imzalanır ve `COSIGN_PUBLIC_KEY` tanımlıysa pipeline içinde doğrulanır; imza dosyaları GitHub Release’e `.sig` olarak yüklenir
* Admin panel IP allowlist
* Config URL’leri **tek kullanım** ve **24 saat TTL**
* Gizli anahtarlar: SOPS/age veya Vault
* Container imzalama (cosign) ve SBOM (syft/grype) — v1.1
* **Sızıntı önleme:** DNS/IPv6 leak testleri, kill‑switch yönergeleri

---

## KVKK/GDPR ve Gizlilik

* **Saklananlar:** hesap, plan/ödeme meta, cihaz public key, bağlantı toplam bayt ve zaman damgaları
* **Saklanmayanlar:** ziyaret edilen siteler, içerik, DNS sorgu içerikleri (varsayılan)
* Silme talebi akışı, veri minimizasyonu, amaç sınırlaması

`docs/privacy.md` içinde aydınlatma metinleri ve saklama süreleri örneği yer alır.

---

## Test Stratejisi

* **Birim:** Auth, ödeme webhook’ları, peer CRUD
* **Entegrasyon:** Agent↔API mTLS, config üretimi
* **Saha:** DNS/IPv6 leak, p95 latency, throughput, port blokaj testi
* **Güvenlik:** SAST, container image taraması, secret scanning

### Komut Örnekleri

```bash
cd backend && go test ./...
cd frontend && pnpm test
```

---

## Yayınlama ve DevOps

* **CI:** GitHub Actions → build + test
* **CD:** SSH/Ansible veya container registry’nin çektiği rollout
* Backend için **blue/green** veya **canary** dağıtım

`.github/workflows/` altında örnek pipeline’lar mevcut.

---

## Altyapı Otomasyonu

### Terraform

* VPC, subnetler, internet gateway, güvenlik grupları
* VM ve statik IP sağlama, DNS kayıtları

### Ansible

* Kernel parametreleri, `wireguard-tools`, `iptables`
* Agent binary veya container deploy

Komut örneği:

```bash
cd infra/terraform/envs/prod
terraform init && terraform apply

cd ../../ansible
ansible-playbook -i inventories/prod hosts.yml playbooks/node.yaml
```

---

## Sızıntı ve Yük Testleri

* `scripts/leaktest.sh` ile **DNS/IPv6 leak** testi
* `iperf3` ile throughput ölçümü
* Port kapatma/engelleme senaryoları

---

## Kapasite Planlama

* Eşikler: aktif peer, 5 dakikalık throughput ortalaması, CPU/IRQ ve NIC offload
* Node **drain** akışı: kapasite altına düşen node’a yeni peer ataması durdurulur

---

## Sorun Giderme

* **Handshake yok:** Saat senkronizasyonu (NTP), firewall UDP portu, MTU düşürmeyi deneyin.
* **DNS çözmüyor:** Panelde tanımlı DNS’leri doğrulayın, kill‑switch kurallarını kontrol edin.
* **Düşük hız:** Yakın bölge seçin, NIC offload ve CPU kullanımını izleyin.

---

## Yol Haritası

* **v1.0 (MVP):** Web panel + WireGuard, Stripe/Iyzico, otomatik node sağlama, temel izleme
* **v1.1:** Mobil uygulamalar, tarayıcı eklentisi, DNS filtreleme seçenekleri
* **v1.2:** Obfuscation protokolleri (Hysteria2/TUIC/REALITY)
* **v1.3:** Dedicated IP ve port‑forward

---

## Sık Sorulanlar

**S: Trafik logluyor musunuz?**
C: Hayır. Sadece toplam bayt ve zaman damgası gibi asgari oturum metrikleri.

**S: Kaç cihaz?**
C: Varsayılan 5.

**S: Hangi platform?**
C: Resmî WireGuard uygulamaları + bizim QR/CONF üretimimiz.

---

## Lisans

* Varsayılan: Tescilli. İstersek OSS bileşenlerine uygun lisans eklenir.
# vpn-tridot
