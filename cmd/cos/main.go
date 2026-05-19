package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rokuosan/cos/internal/codexconfig"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "cos:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usage()
	}
	switch args[0] {
	case "codex-config":
		return runCodexConfig(args[1:])
	case "-h", "--help", "help":
		return usage()
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func usage() error {
	fmt.Fprintln(os.Stderr, "usage: cos codex-config read [--path ~/.codex/config.toml]")
	return nil
}

func runCodexConfig(args []string) error {
	if len(args) == 0 {
		return usage()
	}
	switch args[0] {
	case "read":
		return runCodexConfigRead(args[1:])
	default:
		return fmt.Errorf("unknown codex-config command %q", args[0])
	}
}

func runCodexConfigRead(args []string) error {
	fs := flag.NewFlagSet("codex-config read", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	path := fs.String("path", "~/.codex/config.toml", "Codex config path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := codexconfig.Load(*path)
	if err != nil {
		return err
	}
	for _, project := range cfg.Projects() {
		fmt.Printf("%s\t%s\n", project.Path, project.TrustLevel)
	}
	return nil
}
