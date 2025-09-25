import { defaultApiClient } from "@/lib/api";

const STATUS_VARIANTS: Record<string, string> = {
  online: "bg-emerald-500/20 text-emerald-200",
  planned: "bg-amber-500/20 text-amber-200",
  degraded: "bg-rose-500/20 text-rose-200",
};

interface Region {
  name: string;
  status: string;
  capacity: string;
  peers: number;
}

async function fetchRegions() {
  return defaultApiClient.get<Region[]>("/api/regions", { cache: "no-store" });
}

export default async function RegionsPage() {
  const regions = await fetchRegions();

  return (
    <div className="space-y-8">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-50">Regions</h1>
        <p className="text-sm text-slate-400">
          Monitor capacity and rollout state for each WireGuard cluster.
        </p>
      </header>

      <section className="overflow-hidden rounded-xl border border-slate-800 bg-slate-900/80">
        <table className="min-w-full divide-y divide-slate-800 text-sm">
          <thead className="bg-slate-900/70 text-xs uppercase tracking-wide text-slate-400">
            <tr>
              <th className="px-4 py-3 text-left">Region</th>
              <th className="px-4 py-3 text-left">Status</th>
              <th className="px-4 py-3 text-left">Capacity</th>
              <th className="px-4 py-3 text-left">Peers online</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/60 text-slate-200">
            {regions.map((row) => (
              <tr key={row.name} className="transition hover:bg-slate-800/40">
                <td className="px-4 py-3 font-medium text-slate-100">{row.name}</td>
                <td className="px-4 py-3">
                  <span
                    className={[
                      "rounded-full px-2 py-0.5 text-[11px] font-semibold uppercase",
                      STATUS_VARIANTS[row.status] ?? "bg-slate-700/40 text-slate-200",
                    ].join(" ")}
                  >
                    {row.status}
                  </span>
                </td>
                <td className="px-4 py-3 text-slate-300">{row.capacity}</td>
                <td className="px-4 py-3 text-slate-300">{row.peers}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </div>
  );
}
