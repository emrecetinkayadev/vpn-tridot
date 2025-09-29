# ADR 0005 — React Native WireGuard Köprüsü

- Durum: Önerildi (pending review)
- Tarih: 2024-04-06
- Sahip: @ayse (Mobile Lead)

## Sorun
Mobil MVP için React Native bazlı tek bir kod tabanı istiyoruz; ancak WireGuard tüneli işletim sistemi seviyesinde native bileşenler gerektiriyor. JS tarafında tünel kontrolü yapılamaz; köprü tasarımının güvenli ve sürdürülebilir olması gerekiyor.

## Karar
- React Native uygulaması **bare workflow** kullanacak; iOS ve Android projeleri repo içinde yönetilecek.
- Tünel kontrolü için aşağıdaki native modüller uygulanacak:
  - **iOS:** Swift tabanlı `WGManagerModule` → WireGuardKit + `NETunnelProvider` extension.
  - **Android:** Kotlin tabanlı `WireGuardServiceModule` → `wireguard-go` + gomobile.
- JS tarafında `NativeModules.WireGuard` API yüzeyi standartlaştırılacak:
  - `connect(configId: string): Promise<ConnectionState>`
  - `disconnect(): Promise<void>`
  - `getCurrentState(): Promise<ConnectionState>`
  - Event emitter: `stateChanged`, `error`
- Private key ve hassas yapılandırma native katmanda kalacak; JS tarafına yalnızca `configId` ve durum bilgisi dönecek.
- Monorepo içinde `packages/mobile-core` paketi oluşturularak RN köprüsü için TypeScript tipleri ve durum yöneticisi paylaşılacak.

## Gerekçe
- WireGuardKit ve `wireguard-go` özgür lisanslı ve üretimde kendini kanıtlamış bileşenler.
- Bare workflow Expo/EAS + Fastlane pipeline’ına uyumlu, Network Extension gereksinimleriyle çakışmıyor.
- Tek tip JS API yüzeyi, squad içinde platform bağımsız feature geliştirmeyi hızlandırır.
- Güvenlik prensibi: Private key JS sandbox’ına girmiyor, sadece native bellekte işleniyor.

## Alternatifler
1. **Tamamen native uygulamalar:** iOS/Android için ayrı codebase; ancak toplam iş yükü ve bakım maliyeti artar.
2. **Cordova/Capacitor + native plugin:** RN yerine; modern UI/performans gereksinimlerini karşılamıyor.
3. **Resmi WireGuard uygulaması fork’u:** GPLv2 lisansı ve UI sınırlamaları sebebiyle kapalı kaynak hedefiyle uyumsuz.

## Etkiler
- Mobil CI pipeline’ına Go toolchain ve gomobile kurulumu eklenmeli (Android).
- QA sürecinde gerçek cihaz zorunluluğu: VPN servis izinleri simülatörde çalışmaz.
- Güvenlik taramalarında native modüller için ek kod inceleme süreci.

## Takip İşleri
- [ ] RN köprü paketinin TypeScript API taslağı (owner: @aylin) — 2024-04-10.
- [ ] iOS WireGuardKit POC branch’i (owner: @kerem) — 2024-04-12.
- [ ] Android gomobile build script’i (`scripts/android-wireguard.sh`) (owner: @burak) — 2024-04-15.
- [ ] Güvenlik incelemesi (owner: @elif) — köprü modülleri kodlandığında.

