package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"io"
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