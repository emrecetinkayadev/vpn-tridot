const DEVICE_LIMITS = [
  {
    label: "Default plan limit",
    value: "5 peers",
    hint: "Override per subscription tier",
  },
  {
    label: "WireGuard endpoint",
    value: "vpn.tridot.dev",
    hint: "Replace with region-aware hostnames",
  },
];

const TODO_ITEMS = [
  {
    title: "Create device",
    description: "Wizard to register public key or generate server-side keys.",
    status: "todo",
  },
  {
    title: "Rotate keys",
    description: "One-click peer key rotation with audit trail.",
    status: "todo",
  },
  {
    title: "Config delivery",
    description: "Signed one-time download link + QR code.",
    status: "in-progress",
  },
];

export default function DevicesPage() {
  return (
    <div className="space-y-8">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-50">Devices</h1>
        <p className="text-sm text-slate-400">
          Manage customer peers, enforce limits, and surface WireGuard configs. Backend wiring is
          next up.
        </p>
      </header>

      <section className="grid gap-4 sm:grid-cols-2">
        {DEVICE_LIMITS.map((item) => (
          <div key={item.label} className="rounded-xl border border-slate-800 bg-slate-900/80 p-4">
            <p className="text-xs uppercase tracking-wide text-slate-400">{item.label}</p>
            <p className="mt-2 text-xl font-semibold text-slate-100">{item.value}</p>
            <p className="mt-1 text-xs text-slate-500">{item.hint}</p>
          </div>
        ))}
      </section>

      <section className="space-y-4 rounded-xl border border-slate-800 bg-slate-900/70 p-6">
        <h2 className="text-lg font-semibold text-slate-100">Build queue</h2>
        <ul className="space-y-3 text-sm text-slate-300">
          {TODO_ITEMS.map((item) => (
            <li
              key={item.title}
              className="flex items-start justify-between gap-4 rounded-lg border border-slate-800/60 bg-slate-900/80 px-4 py-3"
            >
              <div>
                <p className="font-medium text-slate-100">{item.title}</p>
                <p className="text-xs text-slate-400">{item.description}</p>
              </div>
              <span
                className={[
                  "mt-1 rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase",
                  item.status === "shipped"
                    ? "bg-emerald-500/20 text-emerald-200"
                    : item.status === "in-progress"
                    ? "bg-sky-500/20 text-sky-200"
                    : "bg-slate-700/40 text-slate-200",
                ].join(" ")}
              >
                {item.status}
              </span>
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}
