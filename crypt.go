package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func encodeEnv(dst string, envMap map[string]string) error {
	passkey := getPassPhrase()

	text, err := godotenv.Marshal(envMap)
	if err != nil {
		return fmt.Errorf("failed to marshal environment variables %w", err)
	}

	encryptedData, err := encrypt(text, passkey)
	if err != nil {
		return fmt.Errorf("failed to encrypt environment variables %w", err)
	}

	if err := os.WriteFile(dst, encryptedData, 0644); err != nil {
		return fmt.Errorf("failed to write encrypted data to disk %w", err)
	}
	slog.Debug("Encrypted data written", "path", dst)
	return nil
}

func encrypt(plaintext string, passkey []byte) (res []byte, err error) {
	block, err := aes.NewCipher(createHash(passkey))
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}

	return gcm.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

func decodeEnv(fp string) (envMap map[string]string, err error) {
	encryptedData, err := os.ReadFile(fp)
	if err != nil {
		slog.Error("failed to read encrypted .env file:", "error", err, "path", fp)
		return
	}

	passkey := getPassPhrase()

	decryptedData, err := decrypt(encryptedData, passkey)
	if err != nil {
		slog.Error("decryption failed", "error", err)
		return
	}

	if envMap, err = godotenv.Unmarshal(decryptedData); err != nil {
		slog.Error("failed to parse decrypted data", "error", err)
		return
	}

	return
}

func decrypt(data []byte, passkey []byte) (res string, err error) {
	block, err := aes.NewCipher(createHash(passkey))
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	if len(data) < gcm.NonceSize() {
		return res, fmt.Errorf("encrypted data is too short")
	}
	nonce, cipherB := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, cipherB, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
