# Node Agent Metrikleri ve Health Payload

Bu belge, VPN node agent'ının kontrol düzlemine gönderdiği health raporunu ve Prometheus metriklerini özetler.

Varsayılan Prometheus sunucusu `:9102` adresinde dinler; `AGENT_METRICS_ADDR` ortam değişkeni ile farklı bir adres/port seçebilirsiniz (ör. `127.0.0.1:9200`). Persist edilen durum `AGENT_STATE_DIR` (default `/var/lib/vpn-agent`) altında `peers.json` ve `drain` dosyalarında tutulur.

## Health Raporu

`POST /api/v1/nodes/health` isteğinde gönderilen JSON:

```jsonc
{
  "timestamp": "2024-05-01T12:00:00Z",
  "wireguard": {
    "peer_count": 4,
    "active_peer_count": 3,
    "handshake_ratio": 0.75,
    "last_handshake": "2024-05-01T11:59:40Z",
    "rx_bytes": 123456,
    "tx_bytes": 654321,
    "rx_bps": 64000,
    "tx_bps": 128000,
    "drain": true
  }
}
```

* `peer_count`: Toplam peer sayısı.
* `active_peer_count`: Son `wg` handshakeleri 3 dk içinde olan peer sayısı.
* `handshake_ratio`: `active_peer_count / peer_count` (0 olduğunda 0 döner).
* `rx_bytes` / `tx_bytes`: Kernel'den okunan kümülatif bayt değerleri.
* `rx_bps` / `tx_bps`: Health çağrıları arasındaki delta üzerinden hesaplanan bit/sn throughput.
* `last_handshake`: En yeni handshake zamanı (UTC). Handshake yoksa alan `null` olur.
* `drain`: Node drain modunda (yeni peer kabul etmeme) ise `true` döner. Drain’i açmak için `touch $AGENT_STATE_DIR/drain` yeterlidir; dosyayı silmek drain’i kapatır.

## Prometheus Endpoint

Node agent, yerel HTTP sunucusunda `/metrics` altında aşağıdaki metrikleri sağlar:

| Metrik | Tip | Açıklama |
| --- | --- | --- |
| `node_agent_wireguard_peers` | Gauge | Toplam peer sayısı |
| `node_agent_wireguard_active_peers` | Gauge | Son 3 dk içinde handshake alan peer sayısı |
| `node_agent_wireguard_handshake_ratio` | Gauge | Aktif/Toplam peer oranı |
| `node_agent_wireguard_rx_bytes_total` | Counter | Kümülatif alınan bayt |
| `node_agent_wireguard_tx_bytes_total` | Counter | Kümülatif gönderilen bayt |
| `node_agent_wireguard_rx_throughput_bps` | Gauge | Son health turunda alınan throughput (bit/sn) |
| `node_agent_wireguard_tx_throughput_bps` | Gauge | Son health turunda gönderilen throughput (bit/sn) |
| `node_agent_wireguard_last_handshake` | Gauge | En yeni handshake UNIX zaman damgası |

> İlk ölçümde throughput metrikleri 0 döner; karşılaştırma için en az iki health turu gerekir.

## Scrape Önerileri

* Prometheus `scrape_interval` değerini agent health döngüsü (`AGENT_POLL_INTERVAL`) ile uyumlu tutun.
* Throughput metriklerinin düzgün hesaplanması için agent'ın ardışık health çağrılarının 0'dan büyük delta üretmesi gerekir.
* Counter metrikleri WireGuard state reset durumunda sıfırlanır; exporter reset'leri handle eder ve counter'a yeni değeri ekler.

## Dashboards

Grafana'da aşağıdaki panel kombinasyonu önerilir:

1. **Peers Aktivite**: `node_agent_wireguard_active_peers` ve `node_agent_wireguard_peers` aynı grafikte.
2. **Throughput**: `*_rx_throughput_bps`, `*_tx_throughput_bps` alanları stacked graph.
3. **Handshake Mesafesi**: `node_agent_wireguard_last_handshake` ile `time()` karşılaştırması (`time() - last_handshake`).

Bu doküman, `docs/OBSERVABILITY.md` ile birlikte okunmalıdır.
