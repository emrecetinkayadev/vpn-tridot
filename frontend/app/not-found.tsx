export default function NotFound() {
  return (
    <div className="mx-auto flex h-full max-w-md flex-col items-center justify-center space-y-4 rounded-2xl border border-slate-800 bg-slate-900/80 p-6 text-center">
      <h1 className="text-2xl font-semibold text-slate-50">Sayfa bulunamadı</h1>
      <p className="text-sm text-slate-400">
        Aradığınız kaynak mevcut değil veya taşınmış olabilir.
      </p>
      <a
        className="rounded-lg border border-slate-700 bg-slate-100 px-4 py-2 text-sm font-semibold text-slate-900 transition hover:bg-white"
        href="/"
      >
        Kontrol paneline dön
      </a>
    </div>
  );
}
