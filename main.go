package main

import (
	"os"

	"github.com/mschneider82/nsqbeat/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
