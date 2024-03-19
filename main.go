package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/joho/godotenv"

	gf "github.com/jessevdk/go-flags"
)

type Options struct {
	Encode bool `short:"E" long:"encode" description:"Encrypts input environment variables into a .env.AES file."`
	Decode bool `short:"D" long:"decode" description:"Decods environment variables from an .env.AES file."`

	Inputs []string `short:"i" long:"input" description:"Filepath to the .env, .env.AES files. By default it uses any .env file in the current directory."`
	Output string   `short:"o" long:"output" description:"Filepath to write the encrypted/decrypted environment variables to." default:".env.AES"`

	AdoptGlobal bool `short:"A" long:"adopt" description:"Adopt the current processes environment variables to add to encoding/decoding."`

	Debug bool `short:"d" long:"debug" description:"Show debug information."`
}

func main() {
	opts := &Options{}
	parser := gf.NewParser(opts, gf.Default)
	parser.Usage = "[options] [arguments]\nNote: if there are conflicting or overlapping flags or options, use '--' to separate `[options]` from `[arguments]`."
	parser.Name = "envi"

	rest, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	if opts.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	slog.Debug("Flags parsed", "Options", opts, "Rest", rest)

	if len(os.Args) == 1 {
		fmt.Println("envi - A simple tool to encrypt and decrypt environment variables.")
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	envMap := map[string]string{}
	if opts.AdoptGlobal {
		for _, env := range os.Environ() {
			kv := strings.SplitN(env, "=", 2)
			envMap[kv[0]] = kv[1]
		}
	}

	res, err := readEnvs(opts.Inputs...)
	if err != nil {
		slog.Error("failed to load input files", "error", err)
		os.Exit(1)
	}

	Combine(envMap, res)

	if opts.Encode {
		encodeEnv(opts.Output, envMap)

		os.Exit(0)
	}

	runCMD(envMap, rest)
}

func runCMD(envMap map[string]string, rest []string) {
	cmd := exec.Command(rest[0], rest[1:]...)

	cmd.Env = Flatten(envMap, func(k, v string) string { return fmt.Sprintf("%s=%s", k, v) })

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		slog.Debug("failed to start command", "error", err)
		panic(err)
	}

	err := cmd.Wait()

	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			os.Exit(status.ExitStatus())
		}
	}
}

func readEnvs(paths ...string) (res map[string]string, err error) {
	res = map[string]string{}
	encRes := map[string]string{} // encrypted envs take precedence, so we store them separately and combine them later

	if len(paths) == 0 {
		if _, err := os.Stat(".env.AES"); err == nil {
			paths = append(paths, ".env.AES")
		} else if _, err := os.Stat(".env"); err == nil {
			paths = append(paths, ".env")
		} else {
			return nil, fmt.Errorf("no .env or .env.AES file found in the current directory")
		}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			return nil, fmt.Errorf("no file found %s %w", path, err)
		}

		if strings.HasSuffix(strings.ToLower(path), ".aes") {
			envMap, err := decodeEnv(path)
			if err != nil {
				return nil, fmt.Errorf("failed to decode file %s %w", path, err)
			}
			Combine(encRes, envMap)
			continue
		}

		envMap, err := godotenv.Read(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read .env %s %w", path, err)
		}

		Combine(res, envMap)
	}

	Combine(res, encRes)
	return
}
