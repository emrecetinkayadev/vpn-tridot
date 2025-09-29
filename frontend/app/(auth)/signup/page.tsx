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
        <fieldset className="space-y-3 rounded-lg border border-slate-800 bg-slate-900/70 px-3 py-3 text-xs text-slate-300">
          <legend className="sr-only">Privacy and consent</legend>
          <p className="text-slate-400">
            We process your personal data according to the TriDot VPN Privacy Notice.
            See <a className="text-slate-200 underline" href="https://vpn.tridot.dev/privacy" target="_blank" rel="noreferrer">vpn.tridot.dev/privacy</a> for full details or contact <a className="text-slate-200 underline" href="mailto:privacy@tridot.dev">privacy@tridot.dev</a>.
          </p>
          <label className="flex items-start gap-3 text-left">
            <input
              type="checkbox"
              required
              className="mt-1 h-4 w-4 rounded border-slate-700 bg-slate-900 text-slate-900 focus:ring-slate-500"
            />
            <span>
              <span className="font-semibold text-slate-200">Mandatory:</span> I read the KVKK/GDPR privacy notice and accept processing of my account and billing data for service delivery.
            </span>
          </label>
          <label className="flex items-start gap-3 text-left">
            <input
              type="checkbox"
              className="mt-1 h-4 w-4 rounded border-slate-700 bg-slate-900 text-slate-900 focus:ring-slate-500"
            />
            <span>
              <span className="font-semibold text-slate-200">Optional:</span> I consent to receive product updates and beta invitations via e-mail. You can withdraw consent anytime from the account settings or by e-mailing privacy@tridot.dev.
            </span>
          </label>
        </fieldset>
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
