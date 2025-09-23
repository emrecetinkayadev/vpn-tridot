# Regions & Nodes

## Overview

The regions service manages the list of VPN regions, node inventory, and node health telemetry. Regions are seeded on bootstrap (`TR-IST`, `TR-IZM`, `EU-FRA`, `EU-NL`) and exposed to the frontend for capacity-based recommendations.

## REST API

### `GET /api/v1/regions`
Returns all regions with aggregated capacity score and active node count.

```json
{
  "regions": [
    {
      "region": {
        "id": "...",
        "code": "TR-IST",
        "name": "İstanbul",
        "country_code": "TR",
        "is_active": true
      },
      "capacity_score": 87,
      "active_nodes": 3
    }
  ]
}
```

### `POST /api/v1/nodes/register`
Registers or updates a node. Requires `X-Provision-Token` header.

Payload:
```json
{
  "region_code": "TR-IST",
  "hostname": "ist-1",
  "public_ipv4": "192.0.2.10",
  "public_ipv6": null,
  "public_key": "...",
  "endpoint": "vpn.example.com:51820",
  "tunnel_port": 51820
}
```
Response: `{ "node_id": "UUID" }`

### `POST /api/v1/nodes/health`
Updates node health metrics and recalculates capacity score. Requires `X-Provision-Token` header.

Payload:
```json
{
  "node_id": "UUID",
  "active_peers": 120,
  "cpu_percent": 55,
  "throughput_mbps": 350,
  "packet_loss": 0.01
}
```
Response: `{ "capacity_score": 73 }`

## Capacity Scoring

The backend applies a simple heuristic:

* Base score = 100
* Subtract 4× active peers (max 60)
* Subtract 0.5× CPU usage
* Subtract throughput/100
* Subtract packet loss × 50

Result is clamped between 0 and 100. Scores feed into `/regions` response for frontend recommendations.

## Provision Secrets

* `NODE_PROVISION_TOKEN` must be set in backend environment (see `.env.example`).
* Node agents must send the token via `X-Provision-Token` header on registration and health updates.

## Seed Data

Bootstrap seeds default regions via `regions.Service.SeedDefaultRegions`. Additional regions can be inserted through SQL migrations or admin tooling.
