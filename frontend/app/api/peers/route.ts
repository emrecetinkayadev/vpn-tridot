import { NextRequest, NextResponse } from "next/server";
import { randomUUID } from "crypto";

export type PeerRecord = {
  id: string;
  name: string;
  publicKey: string;
  allowedIps: string;
  endpoint?: string;
  lastHandshake?: string;
};

export const peers: PeerRecord[] = [
  {
    id: "peer-1",
    name: "MacBook Pro",
    publicKey: "QwxDemoPublicKey1",
    allowedIps: "10.20.0.2/32",
    endpoint: "vpn.tridot.dev:51820",
    lastHandshake: new Date().toISOString(),
  },
  {
    id: "peer-2",
    name: "iPhone 15",
    publicKey: "QwxDemoPublicKey2",
    allowedIps: "10.20.0.3/32",
    endpoint: "vpn.tridot.dev:51820",
  },
];

export async function GET() {
  return NextResponse.json(peers);
}

export async function POST(req: NextRequest) {
  const body = await req.json();
  const name: string = body.name ?? "Unnamed device";
  const publicKey: string = body.publicKey ?? "";

  if (!publicKey) {
    return NextResponse.json({ error: "publicKey is required" }, { status: 400 });
  }

  const peer: PeerRecord = {
    id: randomUUID(),
    name,
    publicKey,
    allowedIps: `10.20.0.${peers.length + 2}/32`,
    endpoint: "vpn.tridot.dev:51820",
    lastHandshake: new Date().toISOString(),
  };
  peers.push(peer);

  return NextResponse.json(peer, { status: 201 });
}
