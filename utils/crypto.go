package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
)

type encryptor struct {
	secret     string
	secretHash []byte
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (encryptor encryptor) Encrypt(s []byte) ([]byte, error) {
	block, _ := aes.NewCipher(encryptor.secretHash)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, s, nil), nil
}

func (encryptor encryptor) Decrypt(s []byte) ([]byte, error) {
	key := encryptor.secretHash
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := s[:nonceSize], s[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func MakeEncryptor(secret string) Encryptor {
	return encryptor{secret: secret, secretHash: []byte(createHash(secret))}
}

type Encryptor interface {
	Encrypt(s []byte) ([]byte, error)
	Decrypt(s []byte) ([]byte, error)
}
