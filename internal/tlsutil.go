package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

// GenerateSelfSignedTLS генерирует self-signed сертификат и ключ для TLS
func GenerateSelfSignedTLS() (certPEM, keyPEM []byte) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	serial := big.NewInt(time.Now().UnixNano())
	certTmpl := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"quic-test"},
			CommonName:   "localhost",
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:     []string{"localhost", "127.0.0.1"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}
	certDER, _ := x509.CreateCertificate(rand.Reader, &certTmpl, &certTmpl, &priv.PublicKey, priv)
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return certPEM, keyPEM
}

// GenerateTLSConfig создает TLS конфигурацию для QUIC
func GenerateTLSConfig(noTLS bool) *tls.Config {
	if noTLS {
		// Для режима без TLS используем самоподписанный сертификат
		certPEM, keyPEM := GenerateSelfSignedTLS()
		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			// Fallback к минимальной конфигурации
			return &tls.Config{
				InsecureSkipVerify: true,
				NextProtos:         []string{"quic-test"},
				MinVersion:         tls.VersionTLS12,
			}
		}
		return &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
			NextProtos:         []string{"quic-test"},
			MinVersion:         tls.VersionTLS12,
		}
	}
	
	// Для режима с TLS генерируем сертификат
	certPEM, keyPEM := GenerateSelfSignedTLS()
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		// Fallback к минимальной конфигурации
		return &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"quic-test"},
			MinVersion:         tls.VersionTLS12,
		}
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"quic-test"},
		MinVersion:   tls.VersionTLS12,
	}
} 