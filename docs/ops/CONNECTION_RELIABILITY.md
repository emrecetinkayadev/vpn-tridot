# İlk Hafta Bağlantı Başarısızlık Oranı < %2

## Tanım
Kullanıcı VPN’e katıldıktan sonraki ilk 7 gün içinde karşılaşılan bağlantı hatalarının toplam denemelere oranı.

## Formül
```
failure_rate = failed_sessions / total_sessions
```
- `failed_sessions`: node agent health raporlarında `handshake_failure=1` olarak işaretlenen girişimler
- `total_sessions`: backend `peers.sessions` tablosunda `status=connected|failed`

## İzleme
- Prometheus metriği: `node_agent_wireguard_handshake_failures_total`
- Grafana paneli: “Connection Reliability” grafiği w/ 7-day rolling window
- Alert: `failure_rate > 0.02` → Slack #ops-alerts (warning)

## Aksiyonlar
- Hata loglarını incele (`journalctl -u agent.service`)
- MTU runbook ve port engelleme testleri çalıştır
- Gerekirse kullanıcıya yeni config üret
