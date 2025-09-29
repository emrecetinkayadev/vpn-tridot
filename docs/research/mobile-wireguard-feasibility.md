# Mobil WireGuard Entegrasyonu Fizibilitesi

## Amaç
VPN-MVP mobil uygulamalarında (iOS/Android) WireGuard tünelini React Native tabanlı arayüzle kontrol edebilmek için teknik yaklaşımı, bağımlılıkları ve riskleri netleştirmek.

## Genel Yaklaşım
- **UI ve iş mantığı**: React Native (TypeScript) + ortak `ui-kit`.
- **Tünel kontrolü**: Native katmanda platforma özgü WireGuard bileşenleri; RN ile köprüler.
- **Konfig kaynağı**: Backend API → secure storage → native tünel API.
- **Durum takibi**: Native tarafından yayılan event'ler (bağlı, bağlanıyor, hata) → RN `NativeEventEmitter`.

## iOS
- **Kütüphane**: [WireGuardKit](https://git.zx2c4.com/wireguard-apple/) (Swift). MIT lisansı, App Store uyumlu.
- **Entegrasyon**:
  - Xcode workspace içerisinde Swift paketi olarak eklenir.
  - `NETunnelProvider` extension gerektirir; bu modül Swift/Obj-C tarafında kalır.
  - RN köprüsü `WGManager` Swift objesi → bağlan/ayır, durum.
- **Konfig yükleme**: Backend'den alınan peer config (`privateKey`, `publicKey`, `endpoint`, `dns`) Keychain'e aktarılır.
- **Daemon**: App extension sandbox'ında WireGuardKit `wg-go` binary'sini yönetir; background tünel Apple Network Extension gereksinimlerine uyar.
- **İzinler**: `com.apple.developer.networking.networkextension` entitlement; MDM kayıt/Apple Developer Program paid.
- **Test**: Xcode UI testleri + Detox (simulator) → VPN extension simülasyonu sınırlı; gerçek cihaz gerekecek.

## Android
- **Kütüphane seçenekleri**:
  1. Resmi WireGuard uygulamasındaki `tunnel` modülü (GPLv2) → lisans kısıtı (kapalı kaynak mümkün değil). Üründe kapalı kaynak hedefleniyorsa uygun değil.
  2. `wireguard-go` + `wg-quick` kombinasyonu; Go kodu JNI ile sarılır. Lisans BSD.
  3. Third-party SDK'lar (örn. TunSafe) güvenlik ve sürdürülebilirlik riskli.
- **Önerilen yol**: `wireguard-go` + minimal Kotlin wrapper.
  - Go kodu `gomobile bind` ile AAR üretilir → Kotlin servisinde kullanılır.
  - Kullanıcıdan VPN hizmeti izni `VpnService` ile alınır.
  - Yaşayan servis foreground notification ile koşar (Android 13).
- **Konfig**: SharedPreferences yerine EncryptedSharedPreferences/SQLCipher; private key sadece native bellekte tutulmalı.
- **CI**: Android NDK, Go 1.22, gomobile toolchain.
- **Test**: Instrumentation (Espresso) + fiziki cihazla throughput ölçümü.

## React Native Köprüleri
- `NativeModules.WireGuard` API surface:
  - `connect(configId)` → promise + durum event'i.
  - `disconnect()`
  - `getStatus()`
  - `onStateChanged(listener)`
- Event emitter platformlar arasında simetrik olmalı.
- Bağlantı hataları i18n mesajlarına map edilir.

## Güvenlik
- Private key hiçbir zaman JS alanına taşınmaz.
- Keychain/Keystore erişimi biometrik korumayla opsiyonel kilitlenir.
- Loglarda IP/anahtar bulunmaz; sadece hata kodu.
- Jailbreak/root tespiti ile devreden çıkarma opsiyonu değerlendirilir.

## Build & Dağıtım
- **CI/CD**: Expo EAS + Fastlane ile imzalama; WireGuardKit Network Extension nedeniyle EAS managed mode uygun değil → bare workflow.
- **Çoklu ortam**: Staging/prod yapı flavor'ları; backend endpoint map'i.
- **Beta**: TestFlight / Play Store Internal Testing.

## Riskler & Açık Sorular
1. Android'de GPL türevi kod kullanmama zorunluluğu → gomobile wrapper için ek iş gücü (~2 hafta).
2. Battery impact: Sürekli bağlantı ile RN app yaşam döngüsü; native servis başlatma mantığı netleşmeli.
3. App Store Network Extension inceleme süresi (~2-3 hafta) takvime eklenmeli.
4. QA ortamında gerçek node/peers ile test için staging sertifikaları.
5. Sürüm büyüklüğü: `wireguard-go` binary >5MB; optimizasyon gerekebilir.

## Sonraki Adımlar
1. iOS için WireGuardKit POC: Xcode projede örnek bağlantı (owner: @kerem) — 1 hafta.
2. Android için gomobile wireguard-go derleme zinciri hazırlığı (owner: @kerem + @burak) — 1.5 hafta.
3. RN köprü API taslağının RFC olarak `docs/adr/` altında paylaşılması.
4. QA gereksinimleri & test case listesi `docs/testplans/mobile.md`.
5. Lisans gözden geçirmesi (Legal/Compliance) özellikle Android tarafı için.

