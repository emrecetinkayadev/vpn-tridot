"use client";

import { useEffect, useState } from "react";

import {
  createPeer,
  deletePeer,
  fetchPeerConfig,
  listPeers,
  renamePeer,
  type Peer,
} from "@/lib/peers";

const DEVICE_LIMITS = [
  {
    label: "Default plan limit",
    value: "5 peers",
    hint: "Override per subscription tier",
  },
  {
    label: "WireGuard endpoint",
    value: "vpn.tridot.dev",
    hint: "Replace with region-aware hostnames",
  },
];

export default function DevicesPage() {
  const [peers, setPeers] = useState<Peer[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [qrPeer, setQrPeer] = useState<Peer | null>(null);
  const [qrDataUrl, setQrDataUrl] = useState<string | null>(null);
  const [qrLoading, setQrLoading] = useState(false);

  const [form, setForm] = useState({
    name: "",
    publicKey: "",
  });

  useEffect(() => {
    refreshPeers();
  }, []);

  async function refreshPeers() {
    try {
      setLoading(true);
      setError(null);
      const data = await listPeers();
      setPeers(data);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  }

  async function handleCreate(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!form.publicKey) {
      setError("Public key gerekli");
      return;
    }
    try {
      setError(null);
      await createPeer({ name: form.name, publicKey: form.publicKey });
      setForm({ name: "", publicKey: "" });
      await refreshPeers();
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function handleRename(peer: Peer) {
    const next = prompt("Yeni cihaz adı", peer.name);
    if (!next || next === peer.name) return;
    try {
      await renamePeer(peer.id, next);
      await refreshPeers();
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function handleDelete(peer: Peer) {
    if (!confirm(`${peer.name} cihazını silmek istediğinize emin misiniz?`)) return;
    try {
      await deletePeer(peer.id);
      await refreshPeers();
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function handleDownload(peer: Peer) {
    try {
      const { config } = await fetchPeerConfig(peer.id);
      const blob = new Blob([config], { type: "text/plain" });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `${peer.name || peer.id}.conf`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function handleShowQr(peer: Peer) {
    try {
      setQrLoading(true);
      setError(null);
      const { config } = await fetchPeerConfig(peer.id);
      const qr = await import("qrcode");
      const dataUrl = await qr.toDataURL(config, { width: 240, margin: 1 });
      setQrPeer(peer);
      setQrDataUrl(dataUrl);
    } catch (err) {
      setError((err as Error).message);
      setQrPeer(null);
      setQrDataUrl(null);
    } finally {
      setQrLoading(false);
    }
  }

  function closeQrModal() {
    setQrPeer(null);
    setQrDataUrl(null);
  }

  return (
    <div className="space-y-8">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-50">Devices</h1>
        <p className="text-sm text-slate-400">
          Manage customer peers, enforce limits, and surface WireGuard configs.
        </p>
        {error ? <p className="text-sm text-rose-300">{error}</p> : null}
      </header>

      <section className="grid gap-4 sm:grid-cols-2">
        {DEVICE_LIMITS.map((item) => (
          <div key={item.label} className="rounded-xl border border-slate-800 bg-slate-900/80 p-4">
            <p className="text-xs uppercase tracking-wide text-slate-400">{item.label}</p>
            <p className="mt-2 text-xl font-semibold text-slate-100">{item.value}</p>
            <p className="mt-1 text-xs text-slate-500">{item.hint}</p>
          </div>
        ))}
      </section>

      <section className="space-y-4 rounded-xl border border-slate-800 bg-slate-900/70 p-6">
        <h2 className="text-lg font-semibold text-slate-100">Yeni cihaz ekle</h2>
        <form onSubmit={handleCreate} className="grid gap-4 md:grid-cols-2">
          <div className="space-y-1">
            <label htmlFor="peer-name" className="text-sm font-medium text-slate-200">
              Cihaz adı
            </label>
            <input
              id="peer-name"
              type="text"
              placeholder="Örneğin iPhone"
              className="w-full rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-500/40"
              value={form.name}
              onChange={(event) => setForm((state) => ({ ...state, name: event.target.value }))}
            />
          </div>
          <div className="space-y-1 md:col-span-1">
            <label htmlFor="peer-public-key" className="text-sm font-medium text-slate-200">
              Public key
            </label>
            <input
              id="peer-public-key"
              type="text"
              placeholder="Base64 WireGuard public key"
              className="w-full rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-500/40"
              value={form.publicKey}
              onChange={(event) => setForm((state) => ({ ...state, publicKey: event.target.value }))}
              required
            />
          </div>
          <div className="md:col-span-2 flex justify-end">
            <button
              type="submit"
              className="rounded-lg border border-slate-600 bg-slate-100 px-4 py-2 text-sm font-semibold text-slate-900 transition hover:bg-white"
              disabled={loading}
            >
              {loading ? "Ekleniyor..." : "Cihaz ekle"}
            </button>
          </div>
        </form>
      </section>

      <section className="space-y-4 rounded-xl border border-slate-800 bg-slate-900/70 p-6">
        <header className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-slate-100">Cihaz listesi</h2>
          <span className="text-xs text-slate-400">
            Toplam {peers.length} cihaz
          </span>
        </header>
        {loading ? (
          <p className="text-sm text-slate-400">Cihazlar yükleniyor...</p>
        ) : peers.length === 0 ? (
          <p className="text-sm text-slate-400">Henüz cihaz eklenmemiş.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-slate-800 text-sm">
              <thead className="bg-slate-900/70 text-xs uppercase tracking-wide text-slate-400">
                <tr>
                  <th className="px-4 py-3 text-left">Cihaz</th>
                  <th className="px-4 py-3 text-left">Public Key</th>
                  <th className="px-4 py-3 text-left">Allowed IPs</th>
                  <th className="px-4 py-3 text-left">Son handshake</th>
                  <th className="px-4 py-3 text-right">İşlemler</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60 text-slate-200">
                {peers.map((peer) => (
                  <tr key={peer.id} className="hover:bg-slate-800/40">
                    <td className="px-4 py-3 font-medium text-slate-100">{peer.name}</td>
                    <td className="px-4 py-3 text-xs text-slate-400 break-all">{peer.publicKey}</td>
                    <td className="px-4 py-3 text-xs text-slate-400">{peer.allowedIps}</td>
                    <td className="px-4 py-3 text-xs text-slate-400">
                      {peer.lastHandshake ? new Date(peer.lastHandshake).toLocaleString() : "—"}
                    </td>
                    <td className="px-4 py-3 text-right space-x-2">
                      <button
                        type="button"
                        className="inline-flex items-center rounded border border-slate-700 px-3 py-1 text-xs text-slate-200 transition hover:border-slate-500"
                        onClick={() => handleRename(peer)}
                      >
                        Yeniden adlandır
                      </button>
                      <button
                        type="button"
                        className="inline-flex items-center rounded border border-slate-700 px-3 py-1 text-xs text-slate-200 transition hover:border-slate-500"
                        onClick={() => handleShowQr(peer)}
                        disabled={qrLoading}
                      >
                        {qrLoading && qrPeer?.id === peer.id ? "Hazırlanıyor..." : "QR Göster"}
                      </button>
                      <button
                        type="button"
                        className="inline-flex items-center rounded border border-slate-700 px-3 py-1 text-xs text-slate-200 transition hover:border-slate-500"
                        onClick={() => handleDownload(peer)}
                      >
                        Config indir
                      </button>
                      <button
                        type="button"
                        className="inline-flex items-center rounded border border-rose-500/40 px-3 py-1 text-xs text-rose-200 transition hover:border-rose-400"
                        onClick={() => handleDelete(peer)}
                      >
                        Sil
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {qrPeer && qrDataUrl ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/70 backdrop-blur">
          <div className="space-y-4 rounded-2xl border border-slate-800 bg-slate-900/90 p-6 text-center shadow-xl">
            <h3 className="text-lg font-semibold text-slate-100">{qrPeer.name} için QR kod</h3>
            <img src={qrDataUrl} alt={`${qrPeer.name} WireGuard QR`} className="mx-auto h-48 w-48" />
            <p className="text-xs text-slate-400">Mobil uygulamada QR taratarak cihazı ekleyin.</p>
            <button
              type="button"
              className="rounded-lg border border-slate-600 bg-slate-100 px-4 py-2 text-sm font-semibold text-slate-900 transition hover:bg-white"
              onClick={closeQrModal}
            >
              Kapat
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}
