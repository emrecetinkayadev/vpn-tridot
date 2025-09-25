import { NextRequest, NextResponse } from "next/server";
import type { PeerRecord } from "../../route";

// reuse in-memory peers array by importing via require
const { peers } = require("../../route") as { peers: PeerRecord[] };

export async function PATCH(
  req: NextRequest,
  { params }: { params: { peerId: string } },
) {
  const peer = peers.find((p: PeerRecord) => p.id === params.peerId);
  if (!peer) {
    return NextResponse.json({ error: "peer not found" }, { status: 404 });
  }

  const body = await req.json();
  if (body.name) {
    peer.name = body.name;
  }

  return NextResponse.json(peer);
}

export async function DELETE(
  _req: NextRequest,
  { params }: { params: { peerId: string } },
) {
  const index = peers.findIndex((p: PeerRecord) => p.id === params.peerId);
  if (index === -1) {
    return NextResponse.json({ error: "peer not found" }, { status: 404 });
  }
  peers.splice(index, 1);

  return NextResponse.json({ success: true });
}
