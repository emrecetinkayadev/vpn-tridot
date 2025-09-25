# İzleme Panoları ve Alarmlar

## Prometheus & Grafana
- Docker compose (prometheus, grafana, loki, alertmanager) `deploy/docker/monitoring-compose.yml`
- Dashboard JSON’ları: `ops/grafana/dashboards/{nodes-overview,peers-overview}.json`
- Datasource provisioning: `ops/grafana/provisioning/datasources`

## Alertmanager
- Config: `ops/alertmanager/alertmanager.yml`
- Kurallar: `ops/prometheus/rules/node-health.yml`
  - NodeAgentDown
  - HandshakeErrorSpike
  - WebhookFailureBurst
  - BackendErrorRateHigh

## Durum
- Grafana UI → TriDot VPN klasöründe panolar yükleniyor.
- Alertmanager 9093 portunda çalışıyor.
- Slack entegrasyonu TODO (secrets eklenince).

## Check-list
- [ ] Slack webhooku tanımlandı (ops env).
- [x] Dashboard JSON’ları repo’da.
- [x] Docker compose ile yerel doğrulama yapılabilir.
