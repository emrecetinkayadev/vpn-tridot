const ACCOUNT_SECTIONS = [
  {
    title: "Profile",
    rows: [
      { label: "Owner", value: "TODO: user.email" },
      { label: "2FA", value: "Not configured" },
      { label: "Last login", value: "—" },
    ],
  },
  {
    title: "Usage",
    rows: [
      { label: "Total traffic", value: "0 GiB" },
      { label: "Last session", value: "—" },
      { label: "Active peers", value: "0" },
    ],
  },
];

export default function AccountPage() {
  return (
    <div className="space-y-8">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-50">Account</h1>
        <p className="text-sm text-slate-400">
          Connect the Go API once auth endpoints go live. For now we render placeholders to guide the
          integration.
        </p>
      </header>

      <section className="grid gap-4 md:grid-cols-2">
        {ACCOUNT_SECTIONS.map((section) => (
          <div key={section.title} className="rounded-xl border border-slate-800 bg-slate-900/70 p-5">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-400">
              {section.title}
            </h2>
            <dl className="mt-4 space-y-3 text-sm">
              {section.rows.map((row) => (
                <div key={row.label} className="flex items-center justify-between gap-3">
                  <dt className="text-slate-400">{row.label}</dt>
                  <dd className="font-medium text-slate-100">{row.value}</dd>
                </div>
              ))}
            </dl>
          </div>
        ))}
      </section>
    </div>
  );
}
