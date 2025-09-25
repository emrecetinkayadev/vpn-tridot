import { NextRequest, NextResponse } from "next/server";
import type { PeerRecord } from "../../../route";

const { peers } = require("../../../route") as { peers: PeerRecord[] };

export async function GET(
  _req: NextRequest,
  { params }: { params: { peerId: string } },
) {
  const peer = peers.find((p: PeerRecord) => p.id === params.peerId);
  if (!peer) {
    return NextResponse.json({ error: "peer not found" }, { status: 404 });
  }

  const config = `
[Interface]
Address = ${peer.allowedIps}
PrivateKey = <redacted>
DNS = 1.1.1.1

[Peer]
PublicKey = ${peer.publicKey}
AllowedIPs = 0.0.0.0/0
Endpoint = ${peer.endpoint ?? ""}
PersistentKeepalive = 25
`.trim();

  return NextResponse.json({ config });
}
