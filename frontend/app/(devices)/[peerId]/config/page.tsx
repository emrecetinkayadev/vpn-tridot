type ConfigPageProps = {
  params: Promise<{ peerId: string }>;
};

export default async function PeerConfigPage({ params }: ConfigPageProps) {
  const { peerId } = await params;
  return (
    <main className="mx-auto max-w-xl py-16">
      <h1 className="text-xl font-semibold">Peer Configuration</h1>
      <p className="text-sm text-slate-400">
        TODO: fetch one-time config payload for peer <strong>{peerId}</strong>.
      </p>
    </main>
  );
}
