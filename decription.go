package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"crypto/rand"
	"io"
)

//----------------------- SECURITER A DOUBLER -------------------------------

type EncryptedData struct {
	IV   string
	Data string
	Tag  string
}

func encrypt(message string) (*EncryptedData, error) {
	// 1. Générer la clé SHA-256 (32 bytes)
	key := sha256.Sum256([]byte(os.Getenv("SECRET_KEY")))

	// 2. Créer le block AES
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	// 3. GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 4. Générer IV (nonce) → 12 bytes
	iv := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// 5. Chiffrer
	encrypted := gcm.Seal(nil, iv, []byte(message), nil)

	// 6. Séparer data et auth tag
	tagSize := gcm.Overhead()
	data := encrypted[:len(encrypted)-tagSize]
	tag := encrypted[len(encrypted)-tagSize:]

	// 7. Encode base64url
	return &EncryptedData{
		IV:   base64.RawURLEncoding.EncodeToString(iv),
		Data: base64.RawURLEncoding.EncodeToString(data),
		Tag:  base64.RawURLEncoding.EncodeToString(tag),
	}, nil
}

func decrypt(ivB64,dataB64,tagB64 string) (string, error) {
	secret := os.Getenv("SECRET_KEY") 

	key := sha256.Sum256([]byte(secret))

	iv, err := base64.RawURLEncoding.DecodeString(ivB64)
	if err != nil {
		fmt.Println("Erreur decrypt N-2")
		return "", err
	}

	data, err := base64.RawURLEncoding.DecodeString(dataB64)
	if err != nil {
		fmt.Println("Erreur decrypt N-3")
		return "", err
	}

	tag, err := base64.RawURLEncoding.DecodeString(tagB64)
	if err != nil {
		fmt.Println("Erreur decrypt N-3")
		return "", err
	}

	block, err := aes.NewCipher(key[:])
	if err != nil {
		fmt.Println("Erreur decrypt N-4")
		return "",err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Println("Erreur decrypt N-5")
		return "",err
	}

	cipherText := append(data, tag...)

	plainText, err := gcm.Open(nil,iv,cipherText,nil)
	if err != nil {
		fmt.Println("Erreur decrypt N-6")
		return "", err
	}
	return string(plainText),nil
}


