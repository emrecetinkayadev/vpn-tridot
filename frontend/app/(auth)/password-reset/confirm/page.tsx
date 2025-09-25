export default function PasswordResetConfirmPage() {
  return (
    <div className="mx-auto max-w-sm space-y-6">
      <header className="space-y-2 text-center">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-50">Yeni şifrenizi belirleyin</h1>
        <p className="text-sm text-slate-400">
          E-postanıza gönderilen bağlantı üzerinden geldiniz. Güvenli bir şifre seçin.
        </p>
      </header>
      <form className="space-y-4 rounded-xl border border-slate-800 bg-slate-900/70 p-6">
        <div className="space-y-1">
          <label htmlFor="password" className="text-sm font-medium text-slate-200">
            Yeni şifre
          </label>
          <input
            id="password"
            type="password"
            placeholder="En az 12 karakter"
            className="w-full rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-500/40"
            required
          />
        </div>
        <div className="space-y-1">
          <label htmlFor="confirm" className="text-sm font-medium text-slate-200">
            Şifreyi doğrula
          </label>
          <input
            id="confirm"
            type="password"
            placeholder="Şifrenizi tekrar yazın"
            className="w-full rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-500/40"
            required
          />
        </div>
        <button
          type="submit"
          className="w-full rounded-lg border border-slate-600 bg-slate-100 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:bg-white"
          disabled
        >
          Şifreyi güncelle (yakında)
        </button>
      </form>
      <p className="text-center text-xs text-slate-400">
        Yardıma mı ihtiyacınız var? <a className="text-slate-200 underline" href="mailto:support@tridot.dev">Destek ile iletişime geçin</a>.
      </p>
    </div>
  );
}
