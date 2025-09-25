export type Peer = {
  id: string;
  name: string;
  publicKey: string;
  allowedIps: string;
  endpoint?: string;
  lastHandshake?: string;
};

async function request<T>(input: RequestInfo, init?: RequestInit): Promise<T> {
  const response = await fetch(input, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
      ...(init?.headers ?? {}),
    },
  });

  if (!response.ok) {
    throw new Error(`Request failed with status ${response.status}`);
  }

  return (await response.json()) as T;
}

export function listPeers(): Promise<Peer[]> {
  return request<Peer[]>("/api/peers", { cache: "no-store" });
}

export function createPeer(payload: { name: string; publicKey: string }): Promise<Peer> {
  return request<Peer>("/api/peers", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function renamePeer(id: string, name: string): Promise<Peer> {
  return request<Peer>(`/api/peers/${id}`, {
    method: "PATCH",
    body: JSON.stringify({ name }),
  });
}

export function deletePeer(id: string): Promise<void> {
  return request<void>(`/api/peers/${id}`, {
    method: "DELETE",
  });
}

export async function fetchPeerConfig(id: string): Promise<{ config: string }> {
  return request<{ config: string }>(`/api/peers/config/${id}`);
}
