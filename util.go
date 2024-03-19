package main

import (
	"crypto/aes"
	"crypto/sha256"
	"fmt"
	"log"
	"os"

	"golang.org/x/term"
)

func Combine[K comparable, V any](dst map[K]V, src map[K]V) map[K]V {
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func Flatten[K comparable, V any, T any](m map[K]V, fn func(K, V) T) (res []T) {
	for key, value := range m {
		res = append(res, fn(key, value))
	}
	return
}

func getPassPhrase() []byte {
	fmt.Println("Enter passkey: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalln("error reading passkey", err)
	}
	return password
}

func createHash(key []byte) []byte {
	hash := sha256.Sum256(key)
	return hash[:aes.BlockSize]
}
