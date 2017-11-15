package main

import (
	"crypto/rand"
	"crypto/tls"
	"time"
)

func getTLSConfig() (tls.Config, error) {
	var tlsConfig tls.Config
	cert, err := tls.LoadX509KeyPair("certs/publickey.cer", "certs/private.key")
	if err != nil {
		return tlsConfig, err
	}

	tlsConfig = tls.Config{Certificates: []tls.Certificate{cert}}
	tlsConfig.Time = func() time.Time { return time.Now() }
	tlsConfig.Rand = rand.Reader
	return tlsConfig, nil
}
