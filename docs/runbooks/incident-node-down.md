# Runbook: Node Agent Kapasite Kaybı / Node Down

## Amaç
TriDot VPN node agent’larından birinin down olması halinde müşteri trafiğini korumak, kapasiteyi geri kazanmak ve olayı raporlamak.

## Belirtiler
- Prometheus uyarısı: `NodeAgentDown`
- Grafana panosu: aktif peer sayısında ani düşüş
- PagerDuty/Slack alarmı: “Node agent unreachable”

## İlk Müdahale
1. **Alarmı doğrula**: Prometheus uyarısını ve Loki loglarını incele.
2. **İkincil kontroller**: `node_agent_wireguard_handshake_ratio` metriği 0’a düşmüş mü? `wg show` çıktısı.
3. **Drain mod**: Backend kontrol panelinden node’u “drain” durumuna al (yeni peer kabul etmesin).

## Teşhis Adımları
1. `ssh bastion -> node` erişimini doğrula. `uptime` ve `systemctl status agent.service`.
2. WireGuard interface durumunu kontrol et:
   ```bash
   sudo wg show
   sudo systemctl status wg-quick@wg0
   ```
3. Atlas veya cloud provider console’dan node’un CPU/Memory grafikleri incelenir.
4. `journalctl -u agent.service` loglarında panik/hata var mı.

## Kurtarma
1. **Servis restart**:
   ```bash
   sudo systemctl restart agent.service
   ```
2. WireGuard interface yeniden kurulmalıysa:
   ```bash
   sudo systemctl restart wg-quick@wg0
   ```
3. Node uzun süre ayakta kalamıyorsa, yeni bir node’u Terraform/Ansible ile provision et; eski node’u “out of rotation” işaretle.

## Sonrası
1. Backend’de node’u tekrar `active` durumuna al.
2. Peer kapasitesi normale döndüğünde Prometheus alarmı kapanır.
3. Olayı Slack #ops-kanalına raporla, ticket aç.
4. Eğer müşteri etkilendiyse postmortem sürecini başlat.

## Bilgi Toplama
- `sudo wg show` çıktısı
- `journalctl -u agent.service --since "-2 hours"`
- Prometheus grafikleri

## Koruyucu Önlemler
- Agent servis watchdog `Restart=always`
- Disk doluluk monitoring
- Terraform autoscaling planı
