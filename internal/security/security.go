package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

// GenerateInMemoryCert создает сертификат и ключ, которые живут только в RAM.
func GenerateInMemoryCert() (tls.Certificate, error) {
	// 1. Генерируем приватный RSA ключ
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)

	// 2. Настраиваем шаблон сертификата
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Secret Chat Inc"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24), // Сертификат живет 24 часа

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 3. Создаем сертификат в формате DER (бинарный)
	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)

	// 4. Кодируем в PEM формат (в памяти)
	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	// 5. Возвращаем готовый для TLS объект
	return tls.X509KeyPair(certPem, keyPem)
}