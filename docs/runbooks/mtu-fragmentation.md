# MTU ve Fragmentasyon Testi

## Amaç
VPN tüneli üzerinden gönderilen paketlerin MTU limitlerine takılmadığını ve fragmentasyonun doğru işlendiğini doğrulamak.

## Adımlar
1. **MTU Keşfi**
   ```bash
   ping -M do -s 1400 vpn-gateway.internal
   ping -M do -s 1420 vpn-gateway.internal
   ```
   - Başarılı en büyük paket boyutunu not edin.
2. **WireGuard MTU Ayarı**
   - Backend `wg` konfiginde `MTU=1420` olarak ayarlanmış olmalı.
   - Eğer `ping` 1420’de başarısız, MTU’yu 1380’e düşürüp tekrar deneyin.
3. **Fragmentasyon**
   ```bash
   iperf3 -c <node-ip> -M 1300 -u
   ```
   - Paket kaybı ve jitter’ı izleyin.
4. **Loglar**
   - `journalctl -u agent.service` içinde MTU şikayetleri var mı kontrol edin.

## Sonuç
- MTU değerini kontrol paneline not edin.
- Başarısız senaryoda firewall veya path MTU discovery ayarlarını gözden geçirin.
