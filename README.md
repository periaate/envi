# Envi
Envi allows for storing and running programs with environment variables which are stored in an encrypted file.

## Help
```
envi - A simple tool to encrypt and decrypt environment variables.
Usage:
  envi [options] [arguments]
Note: if there are conflicting or overlapping flags or options, use '--' to separate `[options]` from `[arguments]`.

Application Options:
  -E, --encode  Encrypts input environment variables into a .env.AES file.
  -D, --decode  Decods environment variables from an .env.AES file.
  -i, --input=  Filepath to the .env, .env.AES files. By default it uses any .env file in the current directory.
  -o, --output= Filepath to write the encrypted/decrypted environment variables to. (default: .env.AES)
  -A, --adopt   Adopt the current processes environment variables to add to encoding/decoding.
  -d, --debug   Show debug information.

Help Options:
  -h, --help    Show this help message
```

## Examples

We will be using the [printenv](./cmd/printenv/) program to demonstrate the functionality of `envi`.

Lets say that we have the following .env file
```
AB=CD
NAME=Periaate
```

Now we can encrypt that .env file into an encrypted version by using the `-E` or `--encode` flags:
```
envi -E
>Enter passkey:
>.env.AES file saved successfully.
```
By default encryption tries to read from `./.env`, but we can also specify which file(s) we want it to use.

We can now use `envi` to call `printenv` with our encrypted env file:
```
envi printenv
>Enter passkey:
AB=CD
NAME=Periaate
```

Let's make another .env file `.env-dev`.
```
DEV_KEY=DEV_VAL
NUM=10
AB=DE
```

We can use a specific file with the `-i` or `--input` flag. If we aren't using an encrypted file, no passkey will be asked.
```
envi -i .env-dev printenv
DEV_KEY=DEV_VAL
NUM=10
AB=DE
```

Multiple inputs may be specified. Environment variables from encrypted files take priority over others.
```
envi -i .env-dev -i .env.AES printenv
>Enter passkey:
AB=CD
NAME=Periaate
DEV_KEY=DEV_VAL
NUM=10
```

Multiple inputs may also be specified when encoding. If the same environment variable exists in multiple files, the prior value will be overwritten.
```
envi -i .env -i .env-dev -E
>Enter passkey:
envi printenv
>Enter passkey:
NAME=Periaate
DEV_KEY=DEV_VAL
NUM=10
AB=DE // `.env`s `AB=CD` overwritten with `.env-dev`s `AB=DE`
```

Files can be decrypted with `-D` or `--decrypt` flags.
```
envi -D
>Enter passkey:
NAME=Periaate
DEV_KEY=DEV_VAL
NUM=10
AB=DE
```
