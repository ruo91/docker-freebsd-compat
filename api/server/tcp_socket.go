package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/docker/docker/pkg/listenbuffer"
)

type tlsConfig struct {
	CA          string
	Certificate string
	Key         string
	Verify      bool
}

func tlsConfigFromServerConfig(conf *ServerConfig) *tlsConfig {
	verify := conf.TlsVerify
	if !conf.Tls && !conf.TlsVerify {
		return nil
	}
	return &tlsConfig{
		Verify:      verify,
		Certificate: conf.TlsCert,
		Key:         conf.TlsKey,
		CA:          conf.TlsCa,
	}
}

func NewTcpSocket(addr string, config *tlsConfig, activate <-chan struct{}) (net.Listener, error) {
	l, err := listenbuffer.NewListenBuffer("tcp", addr, activate)
	if err != nil {
		return nil, err
	}
	if config != nil {
		if l, err = setupTls(l, config); err != nil {
			return nil, err
		}
	}
	return l, nil
}

func setupTls(l net.Listener, config *tlsConfig) (net.Listener, error) {
	tlsCert, err := tls.LoadX509KeyPair(config.Certificate, config.Key)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Could not load X509 key pair (%s, %s): %v", config.Certificate, config.Key, err)
		}
		return nil, fmt.Errorf("Error reading X509 key pair (%s, %s): %q. Make sure the key is encrypted.",
			config.Certificate, config.Key, err)
	}
	tlsConfig := &tls.Config{
		NextProtos:   []string{"http/1.1"},
		Certificates: []tls.Certificate{tlsCert},
		// Avoid fallback on insecure SSL protocols
		MinVersion: tls.VersionTLS10,
	}
	if config.CA != "" {
		certPool := x509.NewCertPool()
		file, err := ioutil.ReadFile(config.CA)
		if err != nil {
			return nil, fmt.Errorf("Could not read CA certificate: %v", err)
		}
		certPool.AppendCertsFromPEM(file)
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = certPool
	}
	return tls.NewListener(l, tlsConfig), nil
}
