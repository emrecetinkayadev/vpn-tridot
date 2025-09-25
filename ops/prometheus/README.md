# Prometheus & Grafana Stack

Bu klasör yerel geliştirme için Prometheus yapılandırmasını içerir.

## Kullanım

```bash
cd deploy/docker
docker compose -f monitoring-compose.yml up -d
```

Prometheus 9090, Grafana 3000, Alertmanager 9093, Loki 3100 portundan yayın yapar. `STATUS_PROMETHEUS_URL` için `.env` dosyasında varsayılan olarak `http://localhost:9090` kullanılır.

## Dosya Yapısı

- `prometheus.yml`: Scrape yapılandırması.
- `rules/`: Uyarı kuralları (`node-health.yml`).
- `targets/`: File SD hedefleri (`staging.yml`).
- `../grafana/dashboards`: Grafana panoları otomatik yüklenir.
- `../grafana/provisioning`: Datasource ve dashboard provisioning.
- `../alertmanager`: Alertmanager konfigürasyonu.
- `../loki`: Loki konfigürasyonu ve veri dizini.
