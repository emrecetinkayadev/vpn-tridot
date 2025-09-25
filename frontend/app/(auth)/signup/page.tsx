import { HCaptchaWidget } from "@/components/forms/HCaptchaWidget";

export default function SignupPage() {
  return (
    <div className="mx-auto max-w-md space-y-6">
      <header className="space-y-2 text-center">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-50">Create an account</h1>
        <p className="text-sm text-slate-400">
          Invite-only for the MVP pilot. Fill the form to request onboarding.
        </p>
      </header>
      <form className="space-y-4 rounded-xl border border-slate-800 bg-slate-900/70 p-6">
        <div className="space-y-1">
          <label htmlFor="name" className="text-sm font-medium text-slate-200">
            Full name
          </label>
          <input
            id="name"
            type="text"
            placeholder="Ada Lovelace"
            className="w-full rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-500/40"
          />
        </div>
        <div className="space-y-1">
          <label htmlFor="email" className="text-sm font-medium text-slate-200">
            Work email
          </label>
          <input
            id="email"
            type="email"
            placeholder="founder@startup.dev"
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
            placeholder="At least 12 characters"
            className="w-full rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-500/40"
          />
        </div>
        <div className="rounded-lg border border-slate-800 bg-slate-900/70 px-3 py-2 text-xs text-slate-400">
          By requesting access you agree to our privacy policy and beta support expectations.
        </div>
        <HCaptchaWidget />
        <button
          type="submit"
          className="w-full rounded-lg border border-slate-600 bg-slate-100 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:bg-white"
          disabled
        >
          Request invite (coming soon)
        </button>
      </form>
      <p className="text-center text-xs text-slate-400">
        Already onboarded? <a className="text-slate-200 underline" href="/login">Sign in</a>.
      </p>
    </div>
  );
}
