package main

import (
	"os"
)

func main() {
	for _, env := range os.Environ() {
		println(env)
	}
}
