package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

// GenerateSelfSignedTLS генерирует self-signed сертификат и ключ для TLS
func GenerateSelfSignedTLS() (certPEM, keyPEM []byte) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	serial := big.NewInt(time.Now().UnixNano())
	certTmpl := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{Organization: []string{"quck-test"}},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(24 * time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	certDER, _ := x509.CreateCertificate(rand.Reader, &certTmpl, &certTmpl, &priv.PublicKey, priv)
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return certPEM, keyPEM
} 