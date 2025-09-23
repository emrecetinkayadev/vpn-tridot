# Peers & Device Management

## Overview

Peers represent WireGuard device slots assigned to a user. Each peer is tied to a region+node and has limits enforced per subscription (default 5 devices).

## API

### `GET /api/v1/peers`
Returns all peers for the authenticated user.

### `POST /api/v1/peers`
Creates a peer. Payload example:

```json
{
  "node_id": "UUID",
  "region_id": "UUID",
  "device_name": "My Laptop",
  "client_public_key": "optional",
  "allowed_ips": "10.0.1.2/32",
  "dns_servers": ["1.1.1.1"],
  "keepalive": 25,
  "mtu": 1420
}
```

Response includes the WireGuard config, client private key (if generated server-side), a one-time config token, and a data URI QR code.

### `PATCH /api/v1/peers/:peerID`
Renames a peer (`device_name`).

### `DELETE /api/v1/peers/:peerID`
Removes the peer and frees a device slot.

### `GET /api/v1/peers/usage`
Returns aggregated usage metrics for the user (total traffic, active peer count, last handshake timestamp).

### `GET /api/v1/peers/config/:token`
Returns the single-use configuration (requires login). Tokens expire after 24 hours.

## Internals

* Keys are generated via `wgtypes.GeneratePrivateKey` when the client does not supply one.
* Config tokens are stored in `user_tokens` table with type `peer_config`; metadata contains the rendered config.
* Capacity scoring relies on node health reports (`POST /api/v1/nodes/health`).
* QR codes are PNG data URIs (base64) produced via the `boombuler/barcode/qr` library.

## Future Work

* Integrate subscription device limits dynamically based on plan.
* Support IPv6 address pools and automatic AllowedIPs assignment.
* Provide admin tooling for forced peer revocation.
