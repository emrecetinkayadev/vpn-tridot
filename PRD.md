# Ürün Gereksinim Dokümanı (PRD)

**Ürün:** Tüketici VPN Servisi
**Sahibi:** Emre
**Sürüm:** MVP → v1.0
**Tarih:** 23 Eylül 2025

---

## 1) Özet ve Hedef

* **Amaç:** Türkiye öncelikli pazarda hızlı, güvenilir ve kolay kurulumlu bir VPN servisi sunmak.
* **Değer Önermesi:** 1 dakikanın altında üyelik + anında yapılandırma (config/QR), düşük gecikme, net gizlilik ilkeleri, şeffaf olmayan‑kayıt politikası, sade uygulama.
* **MVP Süresi:** 6–8 hafta hedef.
* **Başarı Ölçütleri (90 gün):**

  * Aktivasyon süresi (ödeme→ilk bağlanma): **< 2 dk** ortalama
  * İlk hafta bağlantı başarısızlık oranı: **< %2**
  * DNS sızıntı testi başarı oranı: **%100** laboratuvar, **>%98** saha
  * İlk ay ücretli dönüşüm: **>%5** (deneme→ücretli)
  * iade/şikayet oranı: **< %3**

---

## 2) Kapsam

### MVP Kapsamı

1. **Hesap/Üyelik:** E‑posta + parola, OAuth (Google/Apple) v1.1’de.
2. **Planlar ve Ödeme:** Aylık/3 Aylık/Yıllık planlar. Stripe (global) + Iyzico (TR).
3. **Sunucu Bölgeleri:** TR‑İzmir/İstanbul + AB (FRA/NL) başlangıç.
4. **Bağlantı Teknolojisi (Data Plane):** WireGuard (w/ UDP).
5. **İstemci Teslimi:**

   * **Web paneli**: Config/QR üretimi; resmî WireGuard uygulamalarıyla kullanım.
   * **Mobil uygulama (opsiyonel MVP+)**: iOS/Android native client v1.1.
6. **Cihaz Yuvaları:** 5 cihaz varsayılan.
7. **Kullanım Görünürlüğü:** Bağlı cihaz sayısı, toplam veri, son bağlantı zamanı.
8. **Gizlilik İlkesi:** Trafik logu yok, yalnızca **oturum meta verisi (toplam byte, bağlanma zamanı)**.
9. **Otomatik Sağlama:** Terraform + Ansible ile node hazırlama; kontrol düzlemi API’den peer ekleme/çıkarma.
10. **Sağlık/İzleme:** Prometheus + Grafana, alerting.
11. **Destek:** Basit biletleme (e‑posta→HelpScout/Freshdesk entegrasyonu).

### Kapsam Dışı (MVP)

* Çok ileri düzey DPI/obfuscation (Hysteria2/TUIC/REALITY) – v1.2
* Dedicated IP, port‑forward – v1.3
* P2P/Seed optimizasyonları – v1.3
* Tarayıcı eklentisi – v1.1

---

## 3) Persona ve Kullanım Senaryoları

* **Persona A: “Hızlı ve Basit”** – Sosyal medya ve iş uygulamalarına güvenli erişim, tek tık config.
* **Persona B: “Gizlilik Odaklı”** – Günlük trafikte iz bırakmama, açık kaynak uygulamalara yakın durma.

**Temel Kullanıcı Akışları**

1. Kayıt → Plan seç → Ödeme → Sunucu seç → QR/config al → WireGuard ile bağlan.
2. Panelden cihaz ekle/kaldır, sunucu değiştirme, plan yükseltme/iptal.

---

## 4) Sistem Mimarisi

### Genel Görünüm

* **Control Plane (Merkez API):** Hesaplar, planlar, ödeme, cihaz/peer yönetimi, kapasite planlama.
* **Node Agent’ları (Her VPN Sunucusunda):** WireGuard arayüz/peer yönetimi, healthcheck, ölçüm push.
* **Data Plane:** WG tünel trafiği, kernel mod (Linux).

```
Kullanıcı ↔ Web/Mobil UI ↔ Backend API ↔ DB/Queue
                                  ↘
                                   Node Agents ↔ WireGuard
```

### Teknoloji Seçimi (MVP için “hızlı + geliştirilebilir”)

* **Frontend:** Next.js 15 (React), TypeScript, Tailwind v4, shadcn/ui, Tanstack Query.
* **Backend (Control Plane):** Go 1.22+ (Fiber/Gin), **wgctrl** ile WG yönetim API’leri, gRPC (iç), REST+JWT (dış).

  * Neden Go: Tek binary, düşük bellek, eşzamanlılık, hazır wgctrl ekosistemi.
* **Node Agent:** Go micro‑daemon; mTLS ile kontrol düzlemine bağlanır.
* **Veritabanı:** PostgreSQL 16 (RDS/managed"); **Redis** (oturum/ratelimit/queue).
* **Mesajlaşma (opsiyonel):** NATS veya Redis Streams (MVP’de Redis yeterli).
* **İzleme:** Prometheus, Grafana, Loki, Alertmanager.
* **Ödeme:** Stripe + Iyzico.
* **E‑posta:** Postmark/SES.
* **Altyapı:** Oracle/Hetzner bare‑metal/VM; **Terraform** (VPC, SG, LB), **Ansible** (paketler, kernel, wg).
* **Konteyner:** Docker + Compose (node’larda), merkez API için Docker.
* **CI/CD:** GitHub Actions → Build/Push → SSH/Ansible deploy; merkezi API için Blue/Green.

### Ağ ve Güvenlik Topolojisi

* Her node: `wg0` arayüzü, `iptables` NAT + kill‑switch; Cloud firewall’da UDP WG port açılır.
* mTLS: Control Plane ↔ Node Agent.
* **DNS:** İç resolver + reklam/tracker filtreleri v1.1.
* **Key Management:**

  * Sunucu WG private key: node’da disk üzerinde 600 izinli.
  * Peer key: **tercihen istemci tarafı** üretim; alternatif, sunucu tarafı üretim + tek seferlik indirme.

---

## 5) Özellik Gereksinimleri

### 5.1 Kimlik ve Hesap

* E‑posta doğrulama, 2FA (TOTP) v1.1.
* Oturum aç/kap, şifre sıfırla, cihaz listesi.

### 5.2 Planlar ve Faturalama

* Plan tanımları: hız sınırsız, adil kullanım yok; cihaz sınırı var.
* Ödeme akışı: Checkout → Webhook → Abonelik aktif → Peer hakkı açılır.
* İptal/durdurma, geri ödeme iş akışı.

### 5.3 Sunucu ve Bölge Yönetimi

* Bölge listesi, **anlık kapasite puanı** (kullanıcıya öneri).
* Otomatik dağıtım: yeni peer eklerken en düşük doluluk node seçimi.

### 5.4 Peer/Config Yönetimi

* Cihaz başına: public key, allowed IPs, endpoint, DNS, MTU, keepalive.
* QR/`*.conf` indirme.
* Cihaz silme/yeniden adlandırma.

### 5.5 İzleme ve Şeffaflık

* Panelde: son bağlantı zamanı, toplam veri.
* Trafik/içerik logu tutulmaz.

### 5.6 Destek

* Yardım merkezi, e‑posta bileti, temel SSS.

---

## 6) Olmayan Kayıt Politikası ve Veri Saklama

* Saklananlar: hesap bilgisi, plan/ödeme meta verisi, cihaz public key, bağlantı **toplam** byte ve zaman damgaları.
* Saklanmayanlar: ziyaret edilen siteler, DNS sorgu içerikleri (varsayılan).
* **KVKK/GDPR:** Veri minimizasyonu, açık rıza metinleri, silme talebi akışı.

---

## 7) Performans ve Güvenilirlik

* Tek node hedefi: **> 2 Gbps** aggregate throughput (modern VM/metal), **p95 < 50 ms** yerel.
* **SLA (MVP):** Best‑effort. v1.0’da 99.9% haftalık hedef.
* Kapasite sinyalleri: aktif peer sayısı, 5m throughput, CPU/IRQ, NIC offload.
* Auto‑healing: health‑fail node drain, yeni peer yerleştirmeyi durdur.

---

## 8) Güvenlik Gereksinimleri

* mTLS + Least‑privilege; node’da sadece agent root yetkisi.
* Config tek‑kullanımlık indirme URL’si, 24 saat sonra maskeleme.
* Ratelimit ve bot koruması (hCaptcha), admin panel IP allowlist.
* Gizli anahtarlar: SOPS/age veya HashiCorp Vault.
* CI/CD imzalı imajlar (cosign) v1.1.
* **Sızıntı Önleme:** DNS‑leak, IPv6‑leak test suite; kill‑switch yönergeleri.

---

## 9) Uyumluluk ve Hukuk

* KVKK aydınlatma, açık rıza, veri işleme politikaları.
* Kayıt ortamı ve saklama süreleri, mahkeme taleplerine yanıt prosedürü.
* Ödeme güvenliği: PCI DSS uyumlu sağlayıcı, biz kart verisi tutmayız.

---

## 10) İzleme, Günlükleme, Telemetri

* **Toplanan metrikler:** CPU/RAM, NIC throughput, aktif peer sayısı, handshake başarısı, p50/p95 latency.
* **Günlükler:** Uygulama hataları, altyapı olayları. Trafik içeriği yok.
* **Alarm:** node down, handshake error spike, webhook failure, kart reddi oranı.

---

## 11) Fiyatlandırma ve Maliyet

* Başlangıç: **₺49‑89/ay** aralığı A/B.
* Maliyet kalemleri: VM/metal, IP, bant genişliği, ödeme komisyonu, e‑posta, destek.
* Hedef brüt marj: **>%60**.

---

## 12) Riskler ve Azaltımlar

* **DPI/Engellemeler:** WG UDP port varyasyonu, farklı endpoint havuzu. v1.2’de QUIC tabanlı alternatif protokoller.
* **Sunucu Kara Liste:** IP havuzu rotasyonu, ASN çeşitliliği.
* **Ödeme Reddetme:** Yerel sağlayıcı çeşitlendirme, tekrar deneme akışları.
* **Aşırı Yük:** Kapasite uyarı eşikleri, otomatik node ekleme playbook’u.

---

## 13) Yol Haritası

* **Sprint 1 (Hafta 1‑2):** Proje iskeleti, auth, temel UI, DB şeması, ödeme sandbox.
* **Sprint 2 (Hafta 3‑4):** Node agent, peer CRUD, config/QR üretimi, Terraform+Ansible.
* **Sprint 3 (Hafta 5‑6):** İzleme, uyarı, biletleme, SSS, hukuki metinler.
* **Sprint 4 (Hafta 7‑8):** Load test, sızıntı testleri, kapalı beta, fiyatlandırma.
* **v1.1:** Mobil uygulamalar, DNS filtre opsiyonları, tarayıcı eklentisi.
* **v1.2:** Obfuscation protokolleri (Hysteria2/TUIC/REALITY).
* **v1.3:** Dedicated IP, port‑forward.

---

## 14) API Tasarım Özeti (Control Plane)

```
POST   /auth/signup
POST   /auth/login
POST   /auth/2fa/verify
GET    /plans
POST   /checkout/session
POST   /webhooks/stripe
GET    /regions
GET    /regions/{id}/capacity
GET    /peers               # list my devices
POST   /peers               # create device (client-side key upload veya server-side gen)
DELETE /peers/{peerId}
GET    /peers/{peerId}/config   # one-time URL + QR
GET    /account/usage
```

### Ör. Config Üretim Kuralları

* AllowedIPs: `0.0.0.0/0, ::/0`
* DNS: Dahili resolver IP
* Keepalive: 25s
* MTU: 1280‑1420 arası cihaz/OS’a göre sezgisel.

---

## 15) Veri Modeli (Özet)

* **users(id, email, pass\_hash, twofa\_secret, created\_at)**
* **subscriptions(id, user\_id, plan\_id, status, renew\_at)**
* **plans(id, name, price, device\_limit)**
* **payments(id, user\_id, provider, external\_id, amount, status, created\_at)**
* **regions(id, name, country\_code)**
* **nodes(id, region\_id, public\_ip, wg\_listen\_port, capacity\_score, status)**
* **peers(id, user\_id, node\_id, pubkey, assigned\_ip4, assigned\_ip6, name, created\_at)**
* **sessions(id, peer\_id, bytes\_up, bytes\_down, last\_handshake\_at, window)**

Index’ler: `peers(user_id)`, `nodes(region_id,status)`, `sessions(peer_id,window)`.

---

## 16) Test Stratejisi

* **Birim:** Auth, ödeme webhooks, peer CRUD.
* **Entegrasyon:** Agent↔API mTLS, config üretimi.
* **Saha:** DNS/IPv6 sızıntı, p95 latency, throughput, port blokaj testi.
* **Güvenlik:** SAST, container image scan, gizli anahtar taraması.

---

## 17) Operasyon Playbook’u

* Node ekleme: Terraform apply → Ansible (kernel + wireguard-tools) → Agent token → Register.
* Incident: node down → drain, statik sayfa durumu, 2 saat içinde yeniden dengeleme.
* Anahtar rotasyonu: 90 günde bir node WG key dönüşümü (kademeli).

---

## 18) Proje Yapısı (Monorepo, FE/BE ayrı dizinler)

```
repo-root/
  frontend/                # Next.js 15, TS, shadcn, Tailwind v4
    app/
    components/
    lib/
    public/
    package.json
  backend/                 # Go API (Gin/Fiber), REST + gRPC (internal)
    cmd/api/
    internal/
      auth/
      billing/
      peers/
      regions/
      nodes/
      storage/
      wg/
    proto/
    go.mod
  node-agent/              # Go daemon running on VPN nodes
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
    k8s/                   # (v1.0+)
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

### Örnek: `compose.node.yaml`

```yaml
services:
  agent:
    image: registry.example.com/vpn/agent:{{ .TAG }}
    network_mode: host
    cap_add: ["NET_ADMIN"]
    volumes:
      - /etc/wireguard:/etc/wireguard
      - /var/log/agent:/var/log/agent
    environment:
      - CONTROL_PLANE_URL=https://cp.example.com
      - AGENT_TOKEN=${AGENT_TOKEN}
```

### Örnek: Backend API Dockerfile

```dockerfile
FROM golang:1.22 AS build
WORKDIR /src
COPY backend/ .
RUN go build -o /out/api ./cmd/api

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/api /api
USER 65532:65532
ENTRYPOINT ["/api"]
```

### Örnek: GitHub Actions (backend CI)

```yaml
name: backend-ci
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: go build ./...
      - run: go test ./...
```

---

## 19) MVP Çıktıları (Done Tanımı)

* Ücretli plan satın alıp 2 dk içinde bağlanabilen en az 100 beta kullanıcı.
* 3 bölge, toplamda 3+ node; p95 latency metrikleri dashboard’da.
* Sızıntı testleri checklist’i geçmiş sürüm etiketi.
* KVKK/GDPR metinleri yayınlanmış web sitesi + durum sayfası.

---

## 20) Sonraki Adımlar

1. Monorepo iskeletini aç.
2. DB şemasını ve ödeme sandbox’ını kur.
3. İlk node’u Terraform+Ansible ile sağla.
4. Agent↔API mTLS kanalını ayağa kaldır.
5. Peer oluşturma→config/QR→bağlantı uçtan uca testi.

