# Mobil Uygulama Test Planı

## Amaç
VPN-MVP iOS ve Android uygulamalarının piyasaya çıkmadan önce fonksiyonel, güvenlik ve performans kriterlerini karşılamasını sağlamak.

## Kapsam
- Platformlar: iOS 16/17 (iPhone 12, 14 Pro), Android 13/14 (Pixel 6, Samsung S22).
- Sürüm kanalları: Staging, Production.
- Uygulama modülleri: Auth, Plan/Ödeme, Peer Yönetimi, WireGuard bağlantısı, Bildirimler.

## Çevre Gereksinimleri
- Staging backend (`api-staging.vpn.example.com`), staging WireGuard nodeları.
- Test kullanıcıları: `qa-ios`, `qa-android`, kredi kartı test tokenları.
- Test sertifikaları: mTLS client, WireGuard key çiftleri.
- Araçlar: Detox, XCTest, Espresso, Charles Proxy, `iperf3`.

## Test Türleri
1. **Fonksiyonel**
   - Auth: Signup/Login, 2FA, şifre sıfırlama.
   - Abonelik: Plan seçimi, App Store/Play Store içi satın alma, kupon.
   - Peer yönetimi: Cihaz ekle/sil/yeniden adlandır, konfig indirme.
   - WireGuard: Bağlan/ayır, kill-switch, otomatik yeniden bağlanma.
   - Bildirimler: Abonelik hatırlatma, bağlantı uyarıları.
2. **Kullanılabilirlik**
   - Onboarding akışı >80% başarı (5 kullanıcı) süresi <3 dk.
   - Dark/Light tema, erişilebilirlik (VoiceOver/TalkBack).
3. **Performans**
   - Bağlanma süresi <4 sn (P95).
   - Throughput testleri: 500 Mbps down / 300 Mbps up hedefi.
   - Battery drain: 30 dk aktif tünelde <8%.
4. **Güvenlik**
   - Secure storage doğrulaması (Tokenlar JS tarafına sızmamalı).
   - Jailbreak/root tespiti.
   - MITM denemelerinde TLS/mTLS doğrulaması.
   - Log denetimi (anahtar/IP yok).
5. **Uyumluluk**
   - App Store Review Guidelines (Network Extension, gizlilik label).
   - Play Store VPN policy, foreground service gereksinimi.

## Test Senaryoları
| ID | Modül | Senaryo | Araç | Sahip |
|----|--------|---------|------|-------|
| MOB-AUTH-001 | Auth | Email signup + 2FA doğrulama | Detox | @cem |
| MOB-AUTH-002 | Auth | Magic link ile login | XCTest UI | @cem |
| MOB-PAY-010 | Ödeme | App Store subscription → backend eşitleme | XCUITest + backend logs | @burak |
| MOB-PAY-011 | Ödeme | Play Store abonelik yenileme | Espresso | @burak |
| MOB-PEER-020 | Peer | Yeni cihaz ekle, config indir | Detox | @aylin |
| MOB-PEER-021 | Peer | Offline modda config kuyruğu | Espresso | @aylin |
| MOB-VPN-030 | WireGuard | iOS bağlan/ayır, kill-switch | XCTest | @kerem |
| MOB-VPN-031 | WireGuard | Android ağ geçişi (Wi-Fi→LTE) | Espresso | @kerem |
| MOB-VPN-032 | WireGuard | İzin reddi sonrası hata mesajı | Detox | @kerem |
| MOB-NOT-040 | Bildirim | Abonelik bitişi push bildirimi | Firebase Test Lab | @aylin |
| MOB-SEC-050 | Güvenlik | Root/Jailbreak tespiti ve çıkış | Manual | @elif |
| MOB-PERF-060 | Performans | iperf3 throughput ölçümü | Manual | @kerem |

## Test Otomasyonu
- CI pipeline'da nightly Detox (staging) + haftalık gerçek cihaz bulutu (Bitrise/Expo EAS).
- WireGuard bağlantı testleri gerçek cihaz gerektirir → manuel/yarı otomatik.

## Çıkış Kriterleri
- Kritik/major bug (sev1/sev2) sayısı 0.
- Fonksiyonel testlerin ≥ %95’i başarılı (retry sonrası).
- Performans metrikleri hedefleri karşılıyor.
- App Store/Play Store gizlilik beyanları güncel.

## Riskler
- Gerçek cihaz slotlarının kısıtlılığı → planlı rezervasyon.
- Network Extension testleri simülatör desteği sınırlı.
- App Store inceleme süresi nedeniyle regresyon testleri sıklaştırılmalı.

## Zaman Çizelgesi
- Haftalık test döngüsü (Pazartesi başlangıç, Perşembe rapor).
- Sürüm dondurma öncesi 2 tam regresyon turu.

## Raporlama
- Sonuçlar `QA Dashboard` (Notion) ve `#mobile-qa` Slack kanalında paylaşılır.
- Fail olan senaryolar için JIRA etiketi `component:mobile`.

