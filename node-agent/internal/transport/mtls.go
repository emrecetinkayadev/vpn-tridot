package transport

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/config"
)

// NewMTLSClient constructs an HTTP client with mutual TLS configured.
func NewMTLSClient(cfg config.MTLSConfig, timeout time.Duration) (*http.Client, error) {
	rootPool := x509.NewCertPool()

	caPEM := cfg.CACert
	if caPEM == "" && cfg.CACertFile != "" {
		bytes, err := ioutil.ReadFile(cfg.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("read ca file: %w", err)
		}
		caPEM = string(bytes)
	}
	if caPEM == "" {
		return nil, errors.New("mtls ca certificate missing")
	}
	if ok := rootPool.AppendCertsFromPEM([]byte(caPEM)); !ok {
		return nil, errors.New("failed to append ca cert")
	}

	certPEM := cfg.Cert
	if certPEM == "" && cfg.CertFile != "" {
		bytes, err := ioutil.ReadFile(cfg.CertFile)
		if err != nil {
			return nil, fmt.Errorf("read cert file: %w", err)
		}
		certPEM = string(bytes)
	}
	if certPEM == "" {
		return nil, errors.New("client certificate missing")
	}

	keyPEM := cfg.Key
	if keyPEM == "" && cfg.KeyFile != "" {
		bytes, err := ioutil.ReadFile(cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("read key file: %w", err)
		}
		keyPEM = string(bytes)
	}
	if keyPEM == "" {
		return nil, errors.New("client key missing")
	}

	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return nil, fmt.Errorf("load client key pair: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      rootPool,
		MinVersion:   tls.VersionTLS12,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	return client, nil
}
