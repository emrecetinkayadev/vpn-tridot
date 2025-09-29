# Post-MVP Backlog (v1.1–v1.3)

Bu doküman `TODO.md` içindeki **12) Post-MVP** maddelerini ayrıntılandırır. Her ögeye kimlik (PM-xx), hedef sürüm ve kabul kriterleri atandı. Güncellemeleri burada tutarak günlük todo listesini sade bırakıyoruz.

## Yol Haritası Özeti
| Versiyon | Hedef Zaman (MVP yayını +) | Amaç |
| --- | --- | --- |
| v1.1 | 2 ay | Mobil istemciler ve temel platform entegrasyonları |
| v1.2 | 4 ay | Sansürlü ağlar için gizleme/obfuscation katmanı |
| v1.3 | 6 ay | Gelişmiş ağ özellikleri (dedicated IP, port-forward, DNS filtreleri) |

## Öğe Detayları

### PM-01 — Mobil Uygulamalar (iOS & Android)
- **Hedef sürüm:** v1.1
- **Sahip:** Mobile squad (EM: @aylin, iOS: @kerem, Android: @mert)
- **Kapsam:**
  - Native WireGuard SDK entegrasyonu, profil senkronizasyonu API’leri
  - Biometrik kilit ve hızlı bağlantı kısayolu
  - App Store / Google Play dağıtım pipeline’ları
- **Bağımlılıklar:** Backend peer API v1.1, push bildirim servisi seçimi
- **Kabul kriterleri:** Beta kullanıcılarının %90’ı mobil istemci ile bağlantı kurabiliyor, mağaza denetimleri tamamlandı.
- **Riskler:** Mağaza onay süreçleri, Apple NEPacketTunnel uzantısı için ek izinler

### PM-02 — Tarayıcı Eklentisi
- **Hedef sürüm:** v1.1
- **Sahip:** Growth squad (Lead: @deniz)
- **Kapsam:**
  - Chrome/Firefox eklentisi ile WireGuard konfig indirme ve hızlı bağlan linkleri
  - Captive portal ve DNS sızıntı uyarıları
- **Bağımlılıklar:** Backend short-lived config URL API, frontend GraphQL endpoint (opsiyonel)
- **Kabul kriterleri:** Eklenti mağazalarında yayın, bağlantı tetikleme senaryosu QA’dan geçti
- **Riskler:** Tarayıcı API kısıtları, kullanıcı eğitimi

### PM-03 — DNS Filtre Profilleri (Ads / Malware)
- **Hedef sürüm:** v1.3
- **Sahip:** Platform squad (Lead: @okan)
- **Kapsam:**
  - Yönetim panelinden profil seçimi (Temel / Reklam Engelleyici / Kötü Amaçlı Yazılım)
  - Backend’de per-peer DNS listesi ve policy enforcement
  - Node agent tarafında `dnsmasq` veya `blocky` entegrasyonu
- **Bağımlılıklar:** Config API genişlemesi, node agent policy modulü
- **Kabul kriterleri:** Profil değişimi <1 dakika içinde yeni el sıkışmalara yansıyor, leak testleri %100 geçiyor
- **Riskler:** False positive şikayetleri, yönetilebilir blok listesi temini

### PM-04 — Obfuscation Protokolleri (Hysteria2 / TUIC / REALITY)
- **Hedef sürüm:** v1.2
- **Sahip:** Network squad (Lead: @selim)
- **Kapsam:**
  - Pilot node’larda Hysteria2 veya TUIC tünelleri
  - Backend API’de protokol seçimi ve config üretimi
  - Playwright ile bölge bazlı bağlantı testi
- **Bağımlılıklar:** Trafik analizine uygun loglama, ek port açılış izinleri
- **Kabul kriterleri:** Sansürlü ağ testlerinde başarı oranı ≥ %95, throughput düşüşü ≤ %15
- **Riskler:** Hız / gecikme trade-off, karmaşık istemci ayarları

### PM-05 — Dedicated IP & Port Forwarding
- **Hedef sürüm:** v1.3
- **Sahip:** Core networking (Lead: @ece)
- **Kapsam:**
  - Kullanıcı başına statik IPv4/IPv6 tahsisi
  - Self-service port forward UI ve firewall otomasyonu
  - Yeni faturalandırma planı (Add-on) + kullanım ölçümü
- **Bağımlılıklar:** IP havuzu yönetimi, Stripe metered billing
- **Kabul kriterleri:** Port forward isteği <30 sn içinde aktif, abuse izleme kurulmuş
- **Riskler:** IP kıtlığı, kötüye kullanım, regülasyon uyumu

## Takip
- Her PM-xx maddesi için GitHub issue açın ve bu dokümana link verin.
- Sprint planlamasında backlog refinements yapılırken güncelleyin.
- Tamamlanan maddeleri buradan çıkarıp `CHANGELOG.md` içine yansıtın.
