# Help Center — TriDot VPN

## SSS

### TriDot VPN hangi protokolü kullanıyor?
WireGuard tabanlıdır. Node agent mTLS ile kontrol düzlemine bağlanır.

### Kaç cihaz bağlanabilir?
Planınıza göre 5’e kadar cihazı aynı anda kullanabilirsiniz.

### Destek kanalı nedir?
Beta sürecinde `support@tridot.dev` adresinden 24 saat içinde dönüş yapılır.

## Kurulum Rehberleri

### iOS
1. App Store’dan resmi WireGuard uygulamasını indirin.
2. Dashboard’daki cihaz sayfasında `Config indir` butonunu kullanın.
3. QR kodu WireGuard uygulamasıyla taratın.
4. Tüneli aktif edin.

### Android
1. Google Play’den WireGuard uygulamasını yükleyin.
2. Konfig dosyasını telefona indirin.
3. WireGuard açın → `Config import from file or archive` → indirdiğiniz dosyayı seçin.

### macOS
1. WireGuard macOS uygulamasını yükleyin.
2. Dashboard’dan `.conf` dosyasını indirin.
3. WireGuard’da `Import Tunnel(s)` seçeneği ile dosyayı içe aktarın.

### Windows
1. `https://www.wireguard.com/install/` adresinden Windows installer’ı indirin.
2. Uygulamayı başlatın, `Add Tunnel` → `Import from file` ile konfigi içe aktarın.

### Linux
1. `sudo apt install wireguard` (Ubuntu) komutuyla paketleri kurun.
2. Konfig dosyasını `/etc/wireguard/wg0.conf` olarak kaydedin.
3. `sudo wg-quick up wg0` komutuyla tüneli başlatın.

## Sorun Giderme
- DNS sızıntısı kontrolü için `scripts/leaktest.sh` betiğini çalıştırın.
- IPv6 sorunu varsa dashboard’dan IPv6 desteğinin açık olduğunu doğrulayın.
- Destek ekibine logları (`wg show`, sistem saatleri) iletin.
