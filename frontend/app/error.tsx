"use client";

export default function Error({ reset }: { reset: () => void }) {
  return (
    <div className="mx-auto flex h-full max-w-md flex-col items-center justify-center space-y-4 rounded-2xl border border-rose-500/30 bg-rose-500/10 p-6 text-center">
      <h1 className="text-2xl font-semibold text-rose-100">Beklenmedik bir hata oluştu</h1>
      <p className="text-sm text-rose-200/80">
        Sayfayı yenileyerek tekrar deneyin. Sorun devam ederse destek ekibi ile iletişime geçin.
      </p>
      <button
        type="button"
        className="rounded-lg border border-rose-400 bg-rose-300 px-4 py-2 text-sm font-semibold text-rose-950 transition hover:bg-rose-200"
        onClick={reset}
      >
        Yeniden dene
      </button>
    </div>
  );
}
