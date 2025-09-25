import { HCaptchaWidget } from "@/components/forms/HCaptchaWidget";

export default function PasswordResetRequestPage() {
  return (
    <div className="mx-auto max-w-sm space-y-6">
      <header className="space-y-2 text-center">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-50">Şifre sıfırlama</h1>
        <p className="text-sm text-slate-400">
          Giriş e-posta adresinizi girin; sıfırlama bağlantısını gönderelim.
        </p>
      </header>
      <form className="space-y-4 rounded-xl border border-slate-800 bg-slate-900/70 p-6">
        <div className="space-y-1">
          <label htmlFor="email" className="text-sm font-medium text-slate-200">
            E-posta
          </label>
          <input
            id="email"
            type="email"
            placeholder="ornek@firma.com"
            className="w-full rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-500/40"
            required
          />
        </div>
        <HCaptchaWidget />
        <button
          type="submit"
          className="w-full rounded-lg border border-slate-600 bg-slate-100 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:bg-white"
          disabled
        >
          Sıfırlama bağlantısı gönder (yakında)
        </button>
      </form>
      <p className="text-center text-xs text-slate-400">
        Hatırladınız mı? <a className="text-slate-200 underline" href="/login">Giriş yapın</a>.
      </p>
    </div>
  );
}
