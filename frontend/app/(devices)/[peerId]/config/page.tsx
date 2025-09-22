interface ConfigPageProps {
  params: { peerId: string };
}

export default function PeerConfigPage({ params }: ConfigPageProps) {
  return (
    <main className="mx-auto max-w-xl py-16">
      <h1 className="text-xl font-semibold">Peer Configuration</h1>
      <p className="text-sm text-slate-400">
        TODO: fetch one-time config payload for peer <strong>{params.peerId}</strong>.
      </p>
    </main>
  );
}
