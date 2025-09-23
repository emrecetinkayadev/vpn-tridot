# Observability

## Metirikler

* Backend Prometheus endpoint’i default olarak `GET /metrics` altında çalışır.
* Node agent Prometheus endpoint’i `AGENT_METRICS_ADDR` (varsayılan `:9102`) adresinde `/metrics` altında dinler; metrikler için `docs/NODE_AGENT_METRICS.md`.
* Varsayılan namespace: `vpn_backend`, subsystem: `http`.
* Ortam değişkenleri:
  * `METRICS_ENABLED=true` → endpoint açıktır. `false` yaparak kapatabilirsiniz.
  * `METRICS_PATH=/metrics`
  * `METRICS_NAMESPACE=vpn_backend`
  * `METRICS_SUBSYSTEM=http`
* Toplanan metrikler:
  * `http_requests_total{method,path,status}`
  * `http_request_duration_seconds_bucket` (otomatik histogram)

## Loglama

* Uygulama log seviyesi `LOG_LEVEL` ile ayarlanır (`debug|info|warn|error`).
* İstek logları varsayılan olarak açıktır (`LOG_REQUESTS_ENABLED=true`).
* Maskeleme/filtreleme için:
  * `LOG_REQUEST_HEADERS=x-request-id,authorization`
  * `LOG_MASK_HEADERS=authorization,proxy-authorization,x-api-key,cookie,set-cookie`
  * `LOG_REQUEST_QUERY_PARAMS=token`
  * `LOG_MASK_QUERY_PARAMS=token,auth,code,password`
* Maskelenen alanlar log içinde `***` olarak görünür.

## Örnek Konfig

```env
LOG_LEVEL=info
LOG_REQUESTS_ENABLED=true
LOG_REQUEST_HEADERS=x-request-id,authorization
LOG_MASK_HEADERS=authorization,proxy-authorization,x-api-key,cookie,set-cookie
LOG_REQUEST_QUERY_PARAMS=token
LOG_MASK_QUERY_PARAMS=token,auth,code,password

METRICS_ENABLED=true
METRICS_PATH=/metrics
METRICS_NAMESPACE=vpn_backend
METRICS_SUBSYSTEM=http
```

## Test

```bash
# Prometheus endpoint'i lokalde doğrula
curl -s http://localhost:8080/metrics | head
```
