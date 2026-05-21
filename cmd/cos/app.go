package main

import (
	"fmt"
	"io"
)

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return usage(stderr)
	}
	switch args[0] {
	case "sync":
		return runSync(args[1:], stdout, stderr)
	case "codex-config":
		return runCodexConfig(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		return usage(stderr)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func usage(stderr io.Writer) error {
	_, _ = fmt.Fprintln(stderr, "usage: cos <sync|codex-config read> ...")
	return nil
}
