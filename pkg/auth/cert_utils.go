package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/lyft/flytestdlib/logger"
	"github.com/pkg/errors"
	"io/ioutil"
	_ "net/http/pprof" // Required to serve application.

)

func GetSslCredentials(ctx context.Context, certFile, keyFile string) (*x509.CertPool, *tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	logger.Infof(ctx, "Constructing SSL credentials")

	certPool := x509.NewCertPool()
	data, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to read server cert file: %s", certFile)
	}
	if ok := certPool.AppendCertsFromPEM([]byte(data)); !ok {
		return nil, nil, fmt.Errorf("failed to load certificate into the pool")
	}

	return certPool, &cert, nil
}
