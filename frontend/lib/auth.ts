export interface Session {
  userId: string;
  token: string;
  refreshToken: string;
  expiresAt: string;
}

export const authStorageKey = "vpn-mvp-session";

export function loadSession(): Session | null {
  if (typeof window === "undefined") {
    return null;
  }

  const raw = window.localStorage.getItem(authStorageKey);
  return raw ? (JSON.parse(raw) as Session) : null;
}

export function persistSession(session: Session): void {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(authStorageKey, JSON.stringify(session));
}

export function clearSession(): void {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.removeItem(authStorageKey);
}
