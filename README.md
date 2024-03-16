# Envi
Envi allows for storing and running programs with environment variables which are stored in an encrypted file.

## Help
```
envi - A simple tool to encrypt and decrypt environment variables.
Usage:
  envi [flags] [--] [arguments]
Note: if there are overlapping flags, use '--' to separate flags from arguments.

Application Options:
  -E, --encode       Encrypts environment variables into a .env.AES file.
  -A, --adopt        Adopt the current processes environment variables to add to encoding/decoding.
  -a, --add          Use ./.env file to add environment variables to encoding/decoding.
  -i, --input:       Filepath to the .env.AES file. (default: ./.env.AES)
  -f, --file:        Filepath to an .env file. (default: ./.env)
  -e, --env:         Environment variables in the form of key:value. Takes precedence over other environment variables.
  -h, --help         Show this help message.
  -d, --debug        Show debug information.

Help Options:
  -?                Show this help message
  -h, --help         Show this help message
```

## Examples
We will say that we have a program called `printenv` which prints all env variables given to it. We will also have a .env file which looks like this:
```
AB=CD
NAME=Periaate
```

We will first encrypt these to our `.env.AES` file using the `-E` or `--encode` and `-a` or `--add` flag. If we wanted to add our current processes env variables we could use the `-A` or `--adpot` flags to adopt the env vars.
```
envi -Ea
>Enter passkey:
>.env.AES file saved successfully.
```

We can now use `envi` to call `printenv` with our encrypted env file:
```
envi printenv
>Enter passkey:
>"AB=CD"
>"NAME=Periaate"
```

We can overwrite env values by using the `-e` or `--env` flag, which allows us to provide key value pairs in our sub process:
```
envi -e "NAME:Daniel" printenv
>Enter passkey:
>"AB=CD"
>"NAME=Daniel"
```

If rewrote our .env file to look like:
```
AB=12
NEW=FIELD
```

We can now use the `-a` flag when calling printenv:
```
envi -a printenv
>"AB=CD"
>"NAME=Periaate"
>"NEW=FIELD"
```

We can see that the `.env.AES` file takes priority over the `.env` file. Variables given with the `-e` flag take presedence over all others.


If we were calling something which used flags, envi's flags might cause collisions with the flags of the program we are calling. In these cases we can use `--` after having given our flags to not try to parse any flags beyond that.
```
envi -ae "KEY:VAL" -- tool -a
```


## Note
This was implemented from start to finish in a few hours, including the documentation. There are likely bugs with the program, and the examples might be wrong and are not exhaustive by any means.
