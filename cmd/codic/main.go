package main

import (
	"fmt"
	"os"

	"github.com/Jaziel8910/codic/internal/cli"
)

// GitHash is injected at build time via -ldflags "-X main.GitHash=...".
var GitHash = ""

func main() {
	cli.GitHash = GitHash
	if err := cli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
