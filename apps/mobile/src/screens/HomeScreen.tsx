import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  ActivityIndicator,
  Pressable,
  StyleSheet,
  Text,
  View,
  useColorScheme,
} from 'react-native';
import {
  ConnectionEvent,
  TunnelState,
} from '@vpn/mobile-core';
import { wireGuardBridge } from '../bridge/wireguard';

const INITIAL_EVENT: ConnectionEvent = {
  state: 'disconnected',
  timestamp: Date.now(),
  message: 'Ready to connect',
};

const CONNECT_CONFIG_ID = 'demo-config-id';

const HomeScreen: React.FC = () => {
  const [connection, setConnection] = useState<ConnectionEvent>(INITIAL_EVENT);
  const [isBusy, setIsBusy] = useState(false);
  const isDarkMode = useColorScheme() !== 'light';
  const palette = isDarkMode ? darkTheme : lightTheme;
  const isConnected = connection.state === 'connected';
  const isConnecting = connection.state === 'connecting' || isBusy;

  useEffect(() => {
    let mounted = true;

    wireGuardBridge
      .getCurrentState()
      .then((state) => {
        if (!mounted) return;
        setConnection({
          state,
          timestamp: Date.now(),
          message: state === 'disconnected' ? 'Ready to connect' : undefined,
        });
      })
      .catch((error) => {
        console.warn('wireGuardBridge.getCurrentState failed', error);
      });

    const unsubscribe = wireGuardBridge.subscribe((event) => {
      setConnection(event);
      setIsBusy(false);
    });

    return () => {
      mounted = false;
      unsubscribe();
    };
  }, []);

  const handleConnect = useCallback(async () => {
    setIsBusy(true);
    try {
      await wireGuardBridge.connect(CONNECT_CONFIG_ID);
    } catch (error) {
      console.warn('wireGuardBridge.connect failed', error);
      setIsBusy(false);
    }
  }, []);

  const handleDisconnect = useCallback(async () => {
    setIsBusy(true);
    try {
      await wireGuardBridge.disconnect();
    } catch (error) {
      console.warn('wireGuardBridge.disconnect failed', error);
      setIsBusy(false);
    }
  }, []);

  const statusText = useMemo(() => formatState(connection.state), [connection.state]);
  const messageText = connection.message;

  return (
    <View
      style={[styles.root, { backgroundColor: palette.background }]}
    >
      <View
        style={[
          styles.card,
          { backgroundColor: palette.surface, shadowColor: palette.shadow },
        ]}
      >
        <Text style={[styles.heading, { color: palette.heading }]}>WireGuard bağlantısı</Text>
        <Text style={[styles.status, { color: palette.status }]}>{statusText}</Text>
        {messageText ? <Text style={[styles.message, { color: palette.message }]}>{messageText}</Text> : null}
        {isConnecting ? (
          <ActivityIndicator style={styles.spinner} color={palette.accent} />
        ) : null}
        <View style={styles.actions}>
          <PrimaryButton
            label={isConnected ? 'Bağlı' : 'Bağlan'}
            disabled={isConnected || isConnecting}
            onPress={handleConnect}
            palette={palette}
          />
          <SecondaryButton
            label="Bağlantıyı Kes"
            disabled={!isConnected && !isConnecting}
            onPress={handleDisconnect}
            palette={palette}
          />
        </View>
      </View>
    </View>
  );
};

const PrimaryButton: React.FC<{
  label: string;
  disabled?: boolean;
  onPress: () => void;
  palette: Palette;
}> = ({ label, disabled, onPress, palette }) => (
  <Pressable
    accessibilityRole="button"
    disabled={disabled}
    onPress={onPress}
    style={({ pressed }) => [
      styles.primaryButton,
      { backgroundColor: palette.accent },
      disabled && styles.buttonDisabled,
      pressed && !disabled && styles.buttonPressed,
    ]}>
    <Text style={[styles.buttonLabel, { color: palette.onAccent }]}>{label}</Text>
  </Pressable>
);

const SecondaryButton: React.FC<{
  label: string;
  disabled?: boolean;
  onPress: () => void;
  palette: Palette;
}> = ({ label, disabled, onPress, palette }) => (
  <Pressable
    accessibilityRole="button"
    disabled={disabled}
    onPress={onPress}
    style={({ pressed }) => [
      styles.secondaryButton,
      { borderColor: palette.border },
      disabled && styles.buttonDisabled,
      pressed && !disabled && styles.buttonPressed,
    ]}>
    <Text style={[styles.buttonLabel, { color: palette.secondaryText }]}>{label}</Text>
  </Pressable>
);

function formatState(state: TunnelState): string {
  switch (state) {
    case 'connected':
      return 'Durum: Bağlı';
    case 'connecting':
      return 'Durum: Bağlanıyor…';
    case 'error':
      return 'Durum: Hata';
    default:
      return 'Durum: Bağlı değil';
  }
}

type Palette = {
  background: string;
  surface: string;
  shadow: string;
  heading: string;
  status: string;
  message: string;
  accent: string;
  onAccent: string;
  border: string;
  secondaryText: string;
};

const darkTheme: Palette = {
  background: '#020617',
  surface: '#0f172a',
  shadow: '#000',
  heading: '#e2e8f0',
  status: '#cbd5f5',
  message: '#94a3b8',
  accent: '#6366f1',
  onAccent: '#f8fafc',
  border: '#475569',
  secondaryText: '#94a3b8',
};

const lightTheme: Palette = {
  background: '#e2e8f0',
  surface: '#ffffff',
  shadow: '#0f172a33',
  heading: '#1f2937',
  status: '#334155',
  message: '#475569',
  accent: '#4f46e5',
  onAccent: '#f8fafc',
  border: '#cbd5f5',
  secondaryText: '#475569',
};

const styles = StyleSheet.create({
  root: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    paddingHorizontal: 24,
  },
  card: {
    width: '100%',
    maxWidth: 360,
    borderRadius: 16,
    padding: 24,
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.2,
    shadowRadius: 16,
    elevation: 8,
  },
  heading: {
    fontSize: 20,
    fontWeight: '700',
    marginBottom: 8,
  },
  status: {
    fontSize: 16,
    fontWeight: '500',
    marginBottom: 4,
  },
  message: {
    fontSize: 14,
  },
  spinner: {
    marginTop: 16,
  },
  actions: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginTop: 24,
    gap: 12,
  },
  primaryButton: {
    flex: 1,
    borderRadius: 12,
    paddingVertical: 12,
    alignItems: 'center',
  },
  secondaryButton: {
    flex: 1,
    borderRadius: 12,
    paddingVertical: 12,
    alignItems: 'center',
    borderWidth: 1,
  },
  buttonLabel: {
    fontSize: 16,
    fontWeight: '600',
  },
  buttonDisabled: {
    opacity: 0.5,
  },
  buttonPressed: {
    opacity: 0.75,
  },
});

export { HomeScreen };
