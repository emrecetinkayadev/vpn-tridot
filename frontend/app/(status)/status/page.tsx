import { fetchStatusSnapshot } from "@/lib/status";

const STATUS_VARIANTS: Record<string, string> = {
  operational: "bg-emerald-500/20 text-emerald-200",
  degraded: "bg-amber-500/20 text-amber-200",
  outage: "bg-rose-500/20 text-rose-200",
  scheduled: "bg-sky-500/20 text-sky-200",
  resolved: "bg-emerald-500/20 text-emerald-200",
  "in-progress": "bg-sky-500/20 text-sky-200",
  monitoring: "bg-sky-500/20 text-sky-200",
  investigating: "bg-amber-500/20 text-amber-200",
};

type BadgeProps = {
  status: string;
};

function StatusBadge({ status }: BadgeProps) {
  const styles = STATUS_VARIANTS[status] ?? "bg-slate-700/40 text-slate-200";
  return (
    <span className={["rounded-full px-2 py-0.5 text-[11px] font-semibold uppercase", styles].join(" ")}
    >
      {status}
    </span>
  );
}

export default async function StatusPage() {
  const snapshot = await fetchStatusSnapshot();
  const incidentStatus = snapshot.incidents.length === 0 ? "operational" : snapshot.incidents[0]?.status ?? "degraded";

  return (
    <div className="space-y-10">
      <header className="space-y-3 text-center">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-50">Durum panosu</h1>
        <p className="text-sm text-slate-400">
          TriDot VPN servis sağlığının anlık özeti. Prometheus'a bağlandığında metrikler otomatik
          yenilenir.
        </p>
        <p className="text-xs text-slate-500">
          Son güncelleme: <span className="font-semibold text-slate-300">{new Date(snapshot.lastUpdated).toUTCString()}</span>.
          Kaynak: {snapshot.dataSource === "prometheus" ? "Prometheus" : "statik fallback"}
        </p>
        {snapshot.errors ? (
          <p className="text-xs text-amber-300">
            Prometheus'a bağlanırken sorun çıktı: {snapshot.errors.join(" | ")}. Statik veriler gösterildi.
          </p>
        ) : null}
      </header>

      <section className="grid gap-4 lg:grid-cols-3">
        {snapshot.components.map((component) => (
          <div
            key={component.name}
            className="space-y-3 rounded-2xl border border-slate-800/70 bg-slate-900/70 p-5"
          >
            <div className="flex items-center justify-between gap-2">
              <div>
                <p className="text-sm font-semibold text-slate-200">{component.name}</p>
                <p className="text-xs text-slate-400">{component.description}</p>
              </div>
              <StatusBadge status={component.status} />
            </div>
            <div className="space-y-1 text-xs text-slate-400">
              <p>
                Sağlıklı: <span className="font-medium text-slate-200">{new Date(component.since).toUTCString()}</span>
              </p>
              {component.detail ? <p>{component.detail}</p> : null}
            </div>
          </div>
        ))}
      </section>

      <section className="space-y-4 rounded-2xl border border-slate-800/60 bg-slate-900/70 p-6">
        <header className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-slate-100">Aktif olaylar</h2>
            <p className="text-xs text-slate-400">PagerDuty entegrasyonu hazır olduğunda alarm akışı buraya düşer.</p>
          </div>
          <StatusBadge status={incidentStatus} />
        </header>
        {snapshot.incidents.length === 0 ? (
          <p className="rounded-lg border border-slate-800/70 bg-slate-900/70 p-4 text-sm text-slate-300">
            Aktif olay yok. Takipte kalın!
          </p>
        ) : (
          <ul className="space-y-3">
            {snapshot.incidents.map((incident) => (
              <li
                key={incident.id}
                className="rounded-xl border border-slate-800/70 bg-slate-900/70 p-5"
              >
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div>
                    <p className="text-sm font-semibold text-slate-100">{incident.title}</p>
                    <p className="text-xs text-slate-400">{incident.timeline}</p>
                  </div>
                  <StatusBadge status={incident.status} />
                </div>
                <dl className="mt-3 grid gap-2 text-xs text-slate-400 sm:grid-cols-3">
                  <div>
                    <dt className="font-semibold text-slate-300">Incident ID</dt>
                    <dd>{incident.id}</dd>
                  </div>
                  <div>
                    <dt className="font-semibold text-slate-300">Impact</dt>
                    <dd>{incident.impact}</dd>
                  </div>
                  <div>
                    <dt className="font-semibold text-slate-300">Resolution</dt>
                    <dd>{incident.resolution}</dd>
                  </div>
                </dl>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="space-y-4 rounded-2xl border border-slate-800/60 bg-slate-900/70 p-6">
        <header className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-slate-100">Planlı bakım</h2>
          <StatusBadge status="scheduled" />
        </header>
        {snapshot.maintenance.length === 0 ? (
          <p className="text-sm text-slate-300">Takvimde planlı bakım bulunmuyor.</p>
        ) : (
          <ul className="space-y-3">
            {snapshot.maintenance.map((item) => (
              <li
                key={item.id}
                className="rounded-xl border border-slate-800/70 bg-slate-900/70 p-5"
              >
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div>
                    <p className="text-sm font-semibold text-slate-100">{item.title}</p>
                    <p className="text-xs text-slate-400">{item.window}</p>
                  </div>
                  <StatusBadge status={item.status} />
                </div>
                <p className="mt-3 text-xs text-slate-400">{item.scope}</p>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
