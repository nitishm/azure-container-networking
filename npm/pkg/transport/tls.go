package transport

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"google.golang.org/grpc/credentials"
)

const (
	serverCertPEMFilename = "tls.crt"
	serverKeyPEMFilename  = "tls.key"
	caCertPEMFilename     = "ca.crt"
	path                  = "/usr/local/npm"
)

func serverTLSCreds() (credentials.TransportCredentials, error) {
	certFilepath := path + "/" + serverCertPEMFilename
	keyFilepath := path + "/" + serverKeyPEMFilename

	return credentials.NewServerTLSFromFile(certFilepath, keyFilepath)
}

func clientTLSConfig() (*tls.Config, error) {
	caCertFilepath := path + "/" + caCertPEMFilename
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := ioutil.ReadFile(caCertFilepath)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Create the credentials and return it
	return &tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: false,
	}, nil
}
