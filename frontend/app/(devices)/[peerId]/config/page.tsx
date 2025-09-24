type ConfigPageProps = {
  params: Promise<{ peerId: string }>;
};

export default async function PeerConfigPage({ params }: ConfigPageProps) {
  const { peerId } = await params;
  return (
    <div className="mx-auto max-w-lg space-y-4">
      <h1 className="text-2xl font-semibold tracking-tight text-slate-50">Peer configuration</h1>
      <p className="text-sm text-slate-400">
        Fetch a signed, one-time WireGuard config payload for peer <span className="font-semibold text-slate-200">{peerId}</span>.
      </p>
      <div className="space-y-3 rounded-xl border border-slate-800 bg-slate-900/70 p-5 text-sm text-slate-300">
        <p>Integration tasks:</p>
        <ul className="list-disc space-y-1 pl-5 text-xs text-slate-400">
          <li>Call control plane for config blob &amp; temporary download URL.</li>
          <li>Render QR + .conf contents with proper secrets masking.</li>
          <li>Expire link after 24h or once fetched by the client.</li>
        </ul>
      </div>
    </div>
  );
}
