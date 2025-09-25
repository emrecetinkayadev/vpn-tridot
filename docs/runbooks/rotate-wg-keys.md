# Runbook: WireGuard Anahtarı Rotasyonu

## Amaç
VPN node’larında kullanılan WireGuard public/private anahtar çiftlerini düzenli olarak veya güvenlik incident’ı sonrası rotasyonlamak.

## Hazırlık
1. Backend node agent ile senkron: `agent` servisinin çalıştığını doğrula.
2. Vault/SOPS üzerinden node provizyon tokenı ve mTLS sertifikaları hazır.
3. Bakım sırasında trafiği etkileyebilecek müşteriler bilgilendirildi.

## Adımlar
1. **Drain Mode**: Backend panelinden ilgili node’u drain’e alın; yeni peer bağlanmasın.
2. **Mevcut Peer Kaydı**: `agent` API’den `GET /peers` ile aktif peer listesini alın, yedekleyin.
3. **Anahtar Üretimi**:
   ```bash
   umask 077
   wg genkey | tee /etc/wireguard/wg0.key | wg pubkey > /etc/wireguard/wg0.pub
   ```
4. **Config Güncelleme**:
   - `/etc/wireguard/wg0.conf` dosyasında `[Interface] PrivateKey` değerini yeni anahtarla değiştirin.
   - Backend’e yeni public key’i bildirin (API: `POST /nodes/rotate-key`).
5. **Peer Configleri**: Agent backend’den yeni public key ile güncellenmiş peer konfiglerini çeker (`ApplyPeers`).
6. **Servis Restart**:
   ```bash
   sudo systemctl restart wg-quick@wg0
   sudo systemctl restart agent.service
   ```
7. **Doğrulama**:
   - `wg show` → el sıkışma süreleri güncel mi?
   - Prometheus `node_agent_wireguard_handshake_ratio` ≥ 0.95 olmalı.
8. **Drain Kaldırma**: Node’u `active` durumuna çevirin.

## Rollback
- Yeni anahtar sorun çıkarırsa, yedeklenen eski anahtarı geri yazın, servisleri yeniden başlatın.
- Peer configleri eski public key ile yeniden senkronize olur.

## Otomasyon Önerisi
- `node-agent` için `/rotate-key` endpoint’i implement edilirse script ile tetiklenebilir.
- Rotasyon sonrası 24 saat içinde başarıyı doğrulayan cron ekleyin.

## Notlar
- Anahtar dosyalarını `600` izinleriyle saklayın.
- CI/CD’ye otomatik tetikleyici eklemeden önce runbook manuel kapsanmalıdır.
