import { NextResponse } from "next/server";

const regions = [
  {
    name: "TR-IST",
    status: "online",
    capacity: "High",
    peers: 42,
  },
  {
    name: "TR-IZM",
    status: "online",
    capacity: "Nominal",
    peers: 37,
  },
  {
    name: "EU-FRA",
    status: "planned",
    capacity: "Pending",
    peers: 0,
  },
];

export async function GET() {
  return NextResponse.json(regions, { headers: { "Cache-Control": "no-store" } });
}
