const METRICS = [
  {
    label: "Active peers",
    value: "—",
    hint: "Connect Prometheus to populate",
  },
  {
    label: "Regions online",
    value: "3",
    hint: "IST, IZM, EU",
  },
  {
    label: "Monthly revenue",
    value: "₺0",
    hint: "Stripe & Iyzico pending",
  },
  {
    label: "Support queue",
    value: "0",
    hint: "Helpdesk integration TBD",
  },
];

const RELEASE_CHECKLIST = [
  {
    title: "Backend auth & billing",
    items: [
      { label: "Stripe + Iyzico webhooks", status: "blocked" },
      { label: "Peer session limits", status: "todo" },
      { label: "Audit/metrics telemetry", status: "todo" },
    ],
  },
  {
    title: "Infrastructure",
    items: [
      { label: "Terraform staging apply", status: "todo" },
      { label: "Ansible node bootstrap", status: "todo" },
      { label: "Agent drain mode test", status: "todo" },
    ],
  },
];

const FEED_ITEMS = [
  {
    title: "WireGuard provisioning",
    status: "Shipped",
    body: "Node agent now advertises health + throughput metrics.",
  },
  {
    title: "Release pipeline",
    status: "In review",
    body: "GH Actions builds backend, agent, and uploads unsigned artifacts.",
  },
  {
    title: "Frontend MVP",
    status: "Draft",
    body: "UI shell and navigation live — hook APIs and flows next.",
  },
];

export default function Home() {
  return (
    <div className="space-y-10">
      <header className="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-slate-800 bg-slate-900/70 px-5 py-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-slate-50">Operations overview</h1>
          <p className="text-sm text-slate-400">
            Snapshot of rollout progress while backend wiring and infra automation land.
          </p>
        </div>
        <span className="rounded-full border border-slate-700 bg-slate-800 px-3 py-1 text-xs font-semibold text-slate-300">
          MVP build status
        </span>
      </header>

      <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {METRICS.map((metric) => (
          <div
            key={metric.label}
            className="rounded-xl border border-slate-800 bg-slate-900/80 p-4 shadow"
          >
            <p className="text-xs uppercase tracking-wide text-slate-400">
              {metric.label}
            </p>
            <p className="mt-2 text-3xl font-semibold text-slate-50">
              {metric.value}
            </p>
            <p className="mt-1 text-xs text-slate-500">{metric.hint}</p>
          </div>
        ))}
      </section>

      <section className="grid gap-6 lg:grid-cols-3">
        <div className="space-y-6 rounded-xl border border-slate-800 bg-slate-900/70 p-6 lg:col-span-2">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-slate-100">Release checklist</h2>
            <span className="rounded-full border border-slate-700 bg-slate-800 px-2 py-0.5 text-xs text-slate-300">
              MVP sprint
            </span>
          </div>
          <div className="grid gap-4 md:grid-cols-2">
            {RELEASE_CHECKLIST.map((group) => (
              <div key={group.title} className="space-y-3">
                <h3 className="text-sm font-semibold text-slate-200">
                  {group.title}
                </h3>
                <ul className="space-y-2 text-sm text-slate-300">
                  {group.items.map((item) => (
                    <li
                      key={item.label}
                      className="flex items-center justify-between rounded-lg border border-slate-800/70 bg-slate-900/80 px-3 py-2"
                    >
                      <span>{item.label}</span>
                      <span
                        className={[
                          "rounded-full px-2 py-0.5 text-xs font-semibold",
                          item.status === "shipped"
                            ? "bg-emerald-500/20 text-emerald-200"
                            : item.status === "blocked"
                            ? "bg-amber-500/20 text-amber-200"
                            : "bg-slate-700/40 text-slate-200",
                        ].join(" ")}
                      >
                        {item.status.toUpperCase()}
                      </span>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </div>

        <div className="space-y-4 rounded-xl border border-slate-800 bg-slate-900/70 p-6">
          <h2 className="text-lg font-semibold text-slate-100">Build feed</h2>
          <ul className="space-y-4">
            {FEED_ITEMS.map((item) => (
              <li key={item.title} className="space-y-2 rounded-lg border border-slate-800/60 bg-slate-900/80 p-4">
                <div className="flex items-center justify-between text-xs uppercase tracking-wide text-slate-400">
                  <span>{item.title}</span>
                  <span className="rounded-full border border-slate-700 bg-slate-800 px-2 py-0.5 text-[10px] text-slate-300">
                    {item.status}
                  </span>
                </div>
                <p className="text-sm text-slate-300">{item.body}</p>
              </li>
            ))}
          </ul>
        </div>
      </section>
    </div>
  );
}
import React from "react";
