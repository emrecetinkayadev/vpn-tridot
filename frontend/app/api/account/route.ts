import { NextResponse } from "next/server";

const account = {
  profile: {
    owner: "ops@tridot.dev",
    twoFactor: "Enabled",
    lastLogin: new Date().toISOString(),
  },
  usage: {
    totalTraffic: "12.4 GiB",
    lastSession: new Date(Date.now() - 3600 * 1000).toISOString(),
    activePeers: 4,
  },
};

export async function GET() {
  return NextResponse.json(account, { headers: { "Cache-Control": "no-store" } });
}
