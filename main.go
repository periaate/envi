package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"golang.org/x/term"

	gf "github.com/jessevdk/go-flags"
)

type Options struct {
	Encode bool `short:"E" long:"encode" description:"Encrypts environment variables into a .env.AES file."`

	AdoptGlobal bool              `short:"A" long:"adopt" description:"Adopt the current processes environment variables to add to encoding/decoding."`
	AddEnv      bool              `short:"a" long:"add" description:"Use ./.env file to add environment variables to encoding/decoding."`
	Input       string            `short:"i" long:"input" description:"Filepath to the .env.AES file." default:"./.env.AES"`
	EnvFile     string            `short:"f" long:"file" description:"Filepath to an .env file." default:"./.env"`
	EnvArgs     map[string]string `short:"e" long:"env" description:"Environment variables in the form of key:value. Takes precedence over other environment variables."`

	Help bool `short:"h" long:"help" description:"Show this help message."`

	Debug bool `short:"d" long:"debug" description:"Show debug information."`
}

func main() {
	opts := &Options{}
	parser := gf.NewParser(opts, gf.Default)

	rest, err := parser.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if opts.Debug {
		fmt.Println("Options:", opts)
		fmt.Println("Rest:", rest)
	}

	parser.Usage = "[flags] [--] [arguments]\nNote: if there are overlapping flags, use '--' to separate flags from arguments."
	parser.Name = "envi"

	if opts.Help || len(os.Args) == 1 {
		fmt.Println("envi - A simple tool to encrypt and decrypt environment variables.")
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	if opts.Encode {
		encodeEnv(opts)
		return
	}

	decodedMap, err := decodeEnv()
	if err != nil {
		fmt.Println("Failed to load environment variables:", err)
		os.Exit(1)
	}
	envMap := getEnvMap(opts)
	for key, value := range decodedMap {
		envMap[key] = value
	}

	envMap = addEnvArgs(envMap, opts.EnvArgs)

	cmd := exec.Command(rest[0], rest[1:]...)

	cmd.Env = mapToArr(envMap)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	err = cmd.Wait()

	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			os.Exit(status.ExitStatus())
		}
	}
}

func addEnvArgs(envMap map[string]string, envArgs map[string]string) map[string]string {
	for key, value := range envArgs {
		envMap[key] = value
	}
	return envMap
}

func getPassPhrase() []byte {
	fmt.Println("Enter passkey: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalln("error reading passkey", err)
	}
	return password
}

func getEnvMap(opts *Options) map[string]string {
	envMap := make(map[string]string)

	if opts.AdoptGlobal {
		for _, env := range os.Environ() {
			kv := strings.SplitN(env, "=", 2)
			envMap[kv[0]] = kv[1]
		}
	}

	if opts.AddEnv {
		f, err := os.ReadFile(opts.EnvFile)
		if err != nil {
			fmt.Println("Failed to read .env file:", err)
			os.Exit(1)
		}
		envFileMap, err := godotenv.Unmarshal(string(f))
		if err != nil {
			fmt.Println("Failed to parse .env file:", err)
			os.Exit(1)
		}
		for key, value := range envFileMap {
			envMap[key] = value
		}
	}

	return envMap
}

func encodeEnv(opts *Options) {
	envMap := getEnvMap(opts)
	envMap = addEnvArgs(envMap, opts.EnvArgs)

	passkey := getPassPhrase()

	text, err := godotenv.Marshal(envMap)
	if err != nil {
		fmt.Println("Failed to marshal environment variables:", err)
		os.Exit(1)
	}

	encryptedData, err := encrypt(text, passkey)
	if err != nil {
		fmt.Println("Encryption failed:", err)
		return
	}

	if err := os.WriteFile(".env.AES", encryptedData, 0644); err != nil {
		fmt.Println("Failed to write encrypted data to disk:", err)
	} else {
		fmt.Println(".env.AES file saved successfully.")
	}
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

func decodeEnv() (envMap map[string]string, err error) {
	encryptedData, err := os.ReadFile(".env.AES")
	if err != nil {
		fmt.Println("Failed to read .env.AES file:", err)
		return
	}

	passkey := getPassPhrase()

	decryptedData, err := decrypt(encryptedData, passkey)
	if err != nil {
		fmt.Println("Decryption failed:", err)
		return
	}

	if envMap, err = godotenv.Unmarshal(decryptedData); err != nil {
		fmt.Println("Failed to parse decrypted data:", err)
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

func createHash(key []byte) []byte {
	hash := sha256.Sum256(key)
	return hash[:aes.BlockSize]
}

func mapToArr(envMap map[string]string) []string {
	var arr []string
	for key, value := range envMap {
		arr = append(arr, fmt.Sprintf("%s=%s", key, value))
	}
	return arr
}
