async function fetchAccount() {
  const baseUrl = process.env.API_BASE_URL ?? process.env.NEXT_PUBLIC_API_BASE ?? "";
  const url = `${baseUrl}/api/account`.replace(/\/\/api/, "/api");
  const response = await fetch(url, { cache: "no-store" });
  if (!response.ok) {
    throw new Error("Account verisi yüklenemedi");
  }
  return response.json() as Promise<{
    profile: {
      owner: string;
      twoFactor: string;
      lastLogin: string;
    };
    usage: {
      totalTraffic: string;
      lastSession: string;
      activePeers: number;
    };
  }>;
}

export default async function AccountPage() {
  const account = await fetchAccount();

  return (
    <div className="space-y-8">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-50">Account</h1>
        <p className="text-sm text-slate-400">
          Kullanıcı profiliniz ve kullanım metrikleriniz burada görünür. Backend API ile senkron.
        </p>
      </header>

      <section className="grid gap-4 md:grid-cols-2">
        <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-5">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-400">Profile</h2>
          <dl className="mt-4 space-y-3 text-sm">
            <div className="flex items-center justify-between gap-3">
              <dt className="text-slate-400">Owner</dt>
              <dd className="font-medium text-slate-100">{account.profile.owner}</dd>
            </div>
            <div className="flex items-center justify-between gap-3">
              <dt className="text-slate-400">2FA</dt>
              <dd className="font-medium text-slate-100">{account.profile.twoFactor}</dd>
            </div>
            <div className="flex items-center justify-between gap-3">
              <dt className="text-slate-400">Last login</dt>
              <dd className="font-medium text-slate-100">
                {new Date(account.profile.lastLogin).toLocaleString()}
              </dd>
            </div>
          </dl>
        </div>

        <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-5">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-400">Usage</h2>
          <dl className="mt-4 space-y-3 text-sm">
            <div className="flex items-center justify-between gap-3">
              <dt className="text-slate-400">Total traffic</dt>
              <dd className="font-medium text-slate-100">{account.usage.totalTraffic}</dd>
            </div>
            <div className="flex items-center justify-between gap-3">
              <dt className="text-slate-400">Last session</dt>
              <dd className="font-medium text-slate-100">
                {new Date(account.usage.lastSession).toLocaleString()}
              </dd>
            </div>
            <div className="flex items-center justify-between gap-3">
              <dt className="text-slate-400">Active peers</dt>
              <dd className="font-medium text-slate-100">{account.usage.activePeers}</dd>
            </div>
          </dl>
        </div>
      </section>
    </div>
  );
}
