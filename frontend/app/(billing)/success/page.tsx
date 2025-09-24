export default function CheckoutSuccessPage() {
  return (
    <div className="space-y-6 text-center">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-50">Payment received</h1>
        <p className="text-sm text-slate-400">
          We will sync with Stripe/Iyzico and activate peers instantly. Hook the real webhook handler
          to drive this view.
        </p>
      </header>
      <div className="mx-auto max-w-md space-y-4 rounded-xl border border-slate-800 bg-slate-900/70 p-6">
        <p className="text-sm text-slate-300">
          Device slots unlock automatically. You can head to the devices page to create your first
          peer.
        </p>
        <div className="flex justify-center gap-3 text-sm text-slate-400">
          <span className="rounded-full border border-slate-700 bg-slate-800 px-3 py-1">Devices</span>
          <span className="rounded-full border border-slate-700 bg-slate-800 px-3 py-1">
            Download config
          </span>
        </div>
      </div>
    </div>
  );
}
