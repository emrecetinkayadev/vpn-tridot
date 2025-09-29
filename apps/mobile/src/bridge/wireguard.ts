import { NativeEventEmitter, NativeModules } from 'react-native';
import {
  ConnectionEvent,
  DEFAULT_BRIDGE_MODULE,
  TunnelState,
  WireGuardBridge,
  WireGuardEvents,
} from '@vpn/mobile-core';

interface NativeWireGuardModule {
  connect(configId: string): Promise<ConnectionEvent>;
  disconnect(): Promise<void>;
  getCurrentState(): Promise<TunnelState>;
}

const nativeModule = NativeModules[DEFAULT_BRIDGE_MODULE] as NativeWireGuardModule | undefined;

if (!nativeModule) {
  throw new Error(`${DEFAULT_BRIDGE_MODULE} native module is not linked`);
}

const eventEmitter = new NativeEventEmitter(nativeModule as any);

const wireGuardBridge: WireGuardBridge = {
  connect: (configId) => nativeModule.connect(configId),
  disconnect: () => nativeModule.disconnect(),
  getCurrentState: () => nativeModule.getCurrentState(),
  subscribe: (listener) => {
    const subscriptions = [
      eventEmitter.addListener(WireGuardEvents.stateChanged, (event: ConnectionEvent) => {
        listener(event);
      }),
      eventEmitter.addListener(WireGuardEvents.error, (event: Partial<ConnectionEvent>) => {
        listener({
          state: 'error',
          timestamp: event.timestamp ?? Date.now(),
          message: event.message ?? 'Unknown error',
        });
      }),
    ];

    return () => {
      subscriptions.forEach((sub) => sub.remove());
    };
  },
};

export { wireGuardBridge };
