package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/netutil"
)

func TestNewAgentAppliesNATAndKillSwitch(t *testing.T) {
	caPEM, certPEM, keyPEM := generateCertBundle(t)
	tempDir := t.TempDir()
	cfg := config.Config{
		ControlPlane: config.ControlPlaneConfig{
			URL:          "https://cp",
			Timeout:      5 * time.Second,
			RegisterPath: "/nodes/register",
			HealthPath:   "/nodes/health",
		},
		Provision: config.ProvisionConfig{Token: "token"},
		MTLS: config.MTLSConfig{
			CACert: caPEM,
			Cert:   certPEM,
			Key:    keyPEM,
		},
		Agent: config.AgentConfig{PollInterval: 1 * time.Second, MetricsAddress: ":0", StateDirectory: tempDir, MaxRetryInterval: 2 * time.Second},
		WireGuard: config.WireGuardConfig{
			InterfaceName:       "wg0",
			ListenPort:          51820,
			AddressCIDR:         "10.0.0.2/32",
			ConfigDirectory:     tempDir,
			EnableNAT:           true,
			EnableKillSwitch:    true,
			PersistentKeepalive: 25,
		},
	}

	var calls [][]string
	restore := netutil.WithCommandRunner(func(name string, args ...string) ([]byte, error) {
		cmd := append([]string{name}, args...)
		calls = append(calls, cmd)
		return nil, nil
	})
	t.Cleanup(restore)

	a, exp, err := newAgent(cfg)
	require.NoError(t, err)
	require.NotNil(t, a)
	require.NotNil(t, exp)
	require.True(t, len(calls) >= 2)
}

func generateCertBundle(t *testing.T) (string, string, string) {
	t.Helper()

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err)

	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})

	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "Node Agent"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	clientDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, caTemplate, &clientKey.PublicKey, caKey)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)})

	return string(caPEM), string(certPEM), string(keyPEM)
}
