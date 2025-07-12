// Package main is the entrypoint of the CLI.
package main

import (
	"fmt"
	"os"

	"github.com/tx3stn/plex2m3u/cmd"
)

func main() {
	code := 0

	defer func() {
		os.Exit(code)
	}()

	if err := cmd.Run(); err != nil {
		code = 1

		fmt.Printf("%s\n", err.Error())
	}
}
