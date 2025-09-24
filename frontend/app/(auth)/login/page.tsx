export default function LoginPage() {
  return (
    <div className="mx-auto max-w-sm space-y-6">
      <header className="space-y-2 text-center">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-50">Welcome back</h1>
        <p className="text-sm text-slate-400">Enter your credentials to access the control panel.</p>
      </header>
      <form className="space-y-4 rounded-xl border border-slate-800 bg-slate-900/70 p-6">
        <div className="space-y-1">
          <label htmlFor="email" className="text-sm font-medium text-slate-200">
            Email
          </label>
          <input
            id="email"
            type="email"
            placeholder="ops@tridot.dev"
            className="w-full rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-500/40"
          />
        </div>
        <div className="space-y-1">
          <label htmlFor="password" className="text-sm font-medium text-slate-200">
            Password
          </label>
          <input
            id="password"
            type="password"
            placeholder="••••••••"
            className="w-full rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-500/40"
          />
        </div>
        <div className="flex items-center justify-between text-xs text-slate-400">
          <label className="inline-flex items-center gap-2">
            <input type="checkbox" className="h-4 w-4 rounded border-slate-700 bg-slate-900" />
            Remember me
          </label>
          <button type="button" className="text-slate-300 transition hover:text-slate-100">
            Forgot password?
          </button>
        </div>
        <button
          type="submit"
          className="w-full rounded-lg border border-slate-600 bg-slate-100 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:bg-white"
          disabled
        >
          Sign in (coming soon)
        </button>
      </form>
      <p className="text-center text-xs text-slate-400">
        Need access? <a className="text-slate-200 underline" href="/signup">Request an invite</a>.
      </p>
    </div>
  );
}
