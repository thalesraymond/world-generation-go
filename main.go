package main

import (
	"os"

	"github.com/thalesraymond/world-generation-go/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
