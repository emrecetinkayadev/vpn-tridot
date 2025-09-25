# Port Engelleme Senaryoları Testi

## Amaç
VPN trafiğinin farklı UDP portlarını kullanarak engelleme (firewall, ISP) senaryolarını test etmek ve fallback mekanizmalarını doğrulamak.

## Ön Koşullar
- Node agent `wg0` default portunu (51820) dinliyor.
- `iptables` veya cloud firewall üzerinde port kısıtları oluşturma yetkisi.

## Test Adımları
1. **Standart Port (51820)**
   - Baseline throughput ölç: `./scripts/loadtest.sh <node-ip> 30`
2. **UDP Port Engelleme**
   - Firewall’da 51820/udp’yi kapat.
   - Node agent fallback port (örn. 51821) açılır mı kontrol et.
   - `wg show` ile peer’lerin yeni portla tekrar handshake yapıp yapmadığını izle.
3. **TCP Failover (opsiyonel)**
   - Eğer TCP fallback varsa, `AllowedIPs` içinde TCP portu tanımla, tunnel’ı yeniden kur.
4. **Kullanıcı Deneyimi**
   - Frontend’de kullanıcıya “UDP blocked” uyarısı gösteriliyor mu?

## İzleme ve Rapor
- Prometheus `node_agent_wireguard_handshake_failures_total` metriklerini takip et.
- Loki loglarında `udp port blocked` benzeri mesajlar.
- Sonuçları `docs/runbooks/leak-and-load-testing.md` altına ek rapor satırı olarak not al.
