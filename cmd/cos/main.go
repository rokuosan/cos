package main

import (
	"flag"
	"fmt"
	"io"
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
	fmt.Fprintln(os.Stderr, "usage: cos codex-config <read|sync> ...")
	return nil
}

func runCodexConfig(args []string) error {
	if len(args) == 0 {
		return usage()
	}
	switch args[0] {
	case "read":
		return runCodexConfigRead(args[1:], os.Stdout)
	case "sync":
		return runCodexConfigSync(args[1:], os.Stdout)
	default:
		return fmt.Errorf("unknown codex-config command %q", args[0])
	}
}

func runCodexConfigRead(args []string, out io.Writer) error {
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
		if _, err := fmt.Fprintf(out, "%s\t%s\n", project.Path, project.TrustLevel); err != nil {
			return err
		}
	}
	return nil
}

func runCodexConfigSync(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("codex-config sync", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	source := fs.String("source", "", "master Codex config path")
	target := fs.String("target", "~/.codex/config.toml", "current Codex config path")
	var preserve preserveRulesFlag
	fs.Var(&preserve, "preserve", "target-owned dotted TOML path to preserve; may be repeated")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *source == "" {
		return fmt.Errorf("--source is required")
	}
	if len(preserve) == 0 {
		preserve = []codexconfig.PreserveRule{"projects.*"}
	}

	sourceDoc, err := codexconfig.LoadDocument(*source)
	if err != nil {
		return err
	}
	targetDoc, err := codexconfig.LoadDocument(*target)
	if err != nil {
		return err
	}

	_, err = io.WriteString(out, codexconfig.Synthesize(sourceDoc, targetDoc, preserve).String())
	return err
}

type preserveRulesFlag []codexconfig.PreserveRule

func (f *preserveRulesFlag) String() string {
	return ""
}

func (f *preserveRulesFlag) Set(value string) error {
	*f = append(*f, codexconfig.PreserveRule(value))
	return nil
}
