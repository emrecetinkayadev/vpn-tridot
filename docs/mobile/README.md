# Mobil Uygulama Başlangıç Rehberi

Bu not, `apps/mobile` altında bulunan React Native projesini geliştirmeye başlamak için gereken adımları özetler.

## Gerekli Araçlar
- Node.js 22 (repo içindeki `tools/node-v22` kullanılabilir)
- pnpm 9+
- Xcode 15+ (iOS)
- Android Studio Iguana (Android) + Android SDK 34
- CocoaPods 1.14+

## Kurulum
```bash
# Root dizinde, Node 22 ile
export PATH="$(pwd)/tools/node-v22/bin:$PATH"
pnpm install
```

Bu komut workspace içindeki `frontend`, `apps/mobile` ve paket bağımlılıklarını (ör. `@vpn/mobile-core`) kurar.

> Not: iOS derlemeleri için CocoaPods, Android için bir JDK kurulmuş olmalıdır. Bu ortamda `pod` ve `java` komutları bulunmadığından otomatik olarak çalıştırılmadı; yerel makinenizde kurulum yaptıktan sonra aşağıdaki komutlarla devam edin:

```bash
cd apps/mobile/ios && pod install
cd apps/mobile/android && ./gradlew tasks
```

## Çalıştırma
```bash
# Metro bundler
cd apps/mobile
pnpm start

# Ayrı terminallerde platform başlatma
pnpm ios    # iOS simulator
pnpm android # Android emulator / cihaz
```

> İlk iOS build'i için `cd ios && pod install` çalıştırmayı unutmayın.
> Bu adım için Mac'te tam Xcode kurulumu gerekiyor. Şu anda ortamda yalnızca Command Line Tools var; Xcode kurulup `sudo xcode-select --switch /Applications/Xcode.app` yapıldıktan sonra komutu tekrar çalıştırın.

## Paylaşılan Paket
- `packages/mobile-core` tipi ve bridge kontratları sunar.
- React Native projesi `@vpn/mobile-core` alias'ı ile TypeScript ve Babel üzerinden erişir.
- Yeni paylaşılan kod eklerken `pnpm --filter @vpn/mobile-core lint` ve `typecheck` komutlarını çalıştırın.

### Bridge Stubları
- iOS tarafında `WireGuard` isimli `RCTEventEmitter` Swift sınıfı bulunur (`ios/VPNMobile/WireGuard.swift`). Şimdilik bağlan/ayır çağrıları sahte olaylar üretir.
- Android tarafında `WireGuardModule` ve `WireGuardPackage` Kotlin dosyaları vardır; `WireGuard` modül adıyla `stateChanged` eventi yayınlar.
- Her iki platform da `WireGuardEvents.stateChanged` ile `ConnectionEvent` sözleşmesini takip eder.
- JS katmanında `src/bridge/wireguard.ts` dosyası NativeModules üzerinden bu köprüyü kullanır; `subscribe` metodu hata olaylarını da `state: 'error'` olarak normalize eder.
- `HomeScreen` (React bileşeni) bu köprü sayesinde demo config kimliği ile bağlan/ayır akışını gösterir.

## Android SDK Kurulumu
Command line tools ve SDK bileşenleri repo içinde kök kullanıcının ev dizinine kuruldu:

```bash
export JAVA_HOME="/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home"
export ANDROID_SDK_ROOT="$HOME/Library/Android/sdk"
sdkmanager "platform-tools" "platforms;android-34" "build-tools;34.0.0"
```

`apps/mobile/android/local.properties` dosyası `sdk.dir=/Users/emre/Library/Android/sdk` olarak güncellenmiştir. Farklı makinede çalışırken bu yolu kendi kullanıcı adınıza göre düzenlemeyi unutmayın.

Gradle doğrulaması: `cd apps/mobile/android && ./gradlew tasks` (JAVA_HOME yukarıdaki gibi).

## Metro & Babel Ayarları
- `metro.config.js` pnpm monorepo yapısını tanıyacak şekilde root `node_modules` klasörünü izler.
- `babel.config.js` içerisinde `@vpn/mobile-core` alias'ı module-resolver ile tanımlıdır; paylaşılan kodu JS tarafında sorunsuz içe aktarmak için yeterlidir.

## Faydalı Komutlar
```bash
pnpm --filter @vpn/mobile-app lint
pnpm --filter @vpn/mobile-app test
pnpm --filter @vpn/mobile-core typecheck
```

## Sonraki Adımlar
- `@vpn/mobile-core` içine WireGuard köprü tiplerini genişletin.
- Native modül köprüleri (iOS/Android) için proje yapılandırmasını bu iskelet üzerine ekleyin.
- Detox / E2E altyapısı için `apps/mobile` içinde `e2e` klasörü oluşturun ve QA planındaki senaryolarla hizalayın.
