package client

import (
	"crypto/tls"

	"go.etcd.io/etcd/pkg/transport"
)

type Config struct {
	CAFile   string
	CertFile string
	KeyFile  string
}

type ETCDConfig struct {
	Endpoints   []string
	TLSConfig   Config
	LeaderElect bool
}

func (c Config) ClientConfig() (*tls.Config, error) {
	if c.CertFile == "" && c.KeyFile == "" && c.CAFile == "" {
		return nil, nil
	}

	info := &transport.TLSInfo{
		CertFile:      c.CertFile,
		KeyFile:       c.KeyFile,
		TrustedCAFile: c.CAFile,
	}
	tlsConfig, err := info.ClientConfig()
	if err != nil {
		return nil, err
	}

	return tlsConfig, nil
}
