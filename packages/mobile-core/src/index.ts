export type TunnelState = 'disconnected' | 'connecting' | 'connected' | 'error';

export const WireGuardEvents = {
  stateChanged: 'stateChanged',
  error: 'error',
} as const;

export type WireGuardEventName = (typeof WireGuardEvents)[keyof typeof WireGuardEvents];

export interface PeerConfigSummary {
  id: string;
  publicKey: string;
  regionId: string;
  createdAt: string;
  lastHandshakeAt?: string;
}

export interface ConnectionEvent {
  state: TunnelState;
  timestamp: number;
  message?: string;
}

export interface ProvisioningRequest {
  peerId: string;
  deviceName: string;
  platform: 'ios' | 'android';
  appVersion: string;
  publicKey: string;
  locale?: string;
}

export interface ProvisioningResponse {
  configId: string;
  peer: PeerConfigSummary;
  tunnel: WireGuardTunnelSpec;
  issuedAt: string;
  expiresAt: string;
}

export interface WireGuardTunnelSpec {
  address: string[];
  dns: string[];
  endpoint: string;
  serverPublicKey: string;
  allowedIPs: string[];
  persistentKeepalive: number;
  mtu?: number;
}

export interface ProvisioningFailure {
  code: 'invalid_peer' | 'quota_exceeded' | 'network_error' | 'unknown';
  message: string;
  retryAfterSeconds?: number;
}

export interface WireGuardBridge {
  connect(configId: string): Promise<ConnectionEvent>;
  disconnect(): Promise<void>;
  getCurrentState(): Promise<TunnelState>;
  subscribe(listener: (event: ConnectionEvent) => void): () => void;
}

export const DEFAULT_BRIDGE_MODULE = 'WireGuard';
