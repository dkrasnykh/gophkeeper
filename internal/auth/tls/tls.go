package tls

import (
	"crypto/tls"
	"google.golang.org/grpc/credentials"
)

// LoadTLSCredentials connection secured by server-side TLS.
// https://dev.to/techschoolguru/how-to-secure-grpc-connection-with-ssl-tls-in-go-4ph
func LoadTLSCredentials(certFile, keyFile string) (credentials.TransportCredentials, error) {
	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
}
