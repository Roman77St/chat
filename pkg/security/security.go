package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"time"
)

// Генерация пары ключей для обмена
func GenerateDHKeys() (*ecdh.PrivateKey, []byte, error) {
	priv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return priv, priv.PublicKey().Bytes(), nil
}

// Вычисление общего секрета и превращение его в 32-байтный ключ для AES
func DeriveKey(priv *ecdh.PrivateKey, remotePub []byte, password string) ([]byte, error) {
	remote, err := ecdh.P256().NewPublicKey(remotePub)
	if err != nil {
		return nil, err
	}
	secret, err := priv.ECDH(remote)
	if err != nil {
		return nil, err
	}

	// Смешиваем DH-секрет с паролем через SHA-256
	h := sha256.New()
	h.Write(secret)
	h.Write([]byte(password)) // Пароль выступает в роли "соли"

	return h.Sum(nil), nil
}

// Шифрование AES-GCM (Authenticated Encryption)
func Encrypt(key, plaintext []byte) ([]byte, error) {
	block, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, aesgcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	// Данные будут: [nonce][ciphertext+tag]
	return aesgcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Дешифрование
func Decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
        return nil, err
    }
	aesgcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
	nonceSize := aesgcm.NonceSize()
	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aesgcm.Open(nil, nonce, actualCiphertext, nil)
}

// GenerateInMemoryCert создает сертификат и ключ, которые живут только в RAM.
func GenerateInMemoryCert() (tls.Certificate, error) {
	// 1. Генерируем ключи Ed25519
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	// 2. Настраиваем шаблон сертификата
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Secret Chat Inc"},
			CommonName:   "localhost",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24), // Сертификат живет 24 часа

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 3. Создаем сертификат в формате DER (бинарный)
	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, pub, priv)

	// 4. Кодируем в PEM формат (в памяти)
	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	// 5. Возвращаем готовый для TLS объект
	return tls.X509KeyPair(certPem, keyPem)
}