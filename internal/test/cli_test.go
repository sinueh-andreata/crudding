package test

import (
	"bytes"
	"strings"
	"testing"

	"crudding/internal/cli"
)

func TestCLIHelpAndUsage(t *testing.T) {
	for _, arg := range []string{"-h", "--help", "help"} {
		var stdout, stderr bytes.Buffer
		code := cli.Run([]string{arg}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("Run(%q) exit code = %d, want 0", arg, code)
		}
		if stderr.Len() != 0 {
			t.Fatalf("Run(%q) unexpected stderr = %q", arg, stderr.String())
		}
		if got := stdout.String(); !strings.Contains(got, "crudding add endpoint") || !strings.Contains(got, "uso:") {
			t.Fatalf("Run(%q) stdout = %q, missing usage entries", arg, got)
		}
	}
}

func TestCLIUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"foo", "bar"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("Run(unknown) exit code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("Run(unknown) unexpected stdout = %q", stdout.String())
	}
	if got := stderr.String(); !strings.Contains(got, "comando desconhecido: foo bar") || !strings.Contains(got, cli.Usage()) {
		t.Fatalf("Run(unknown) stderr = %q", got)
	}
}

func TestCLICommandsRegistry(t *testing.T) {
	cmds := cli.Commands()
	if len(cmds) < 2 {
		t.Fatalf("len(Commands()) = %d, want >= 2", len(cmds))
	}
	for _, cmd := range cmds {
		if cmd.Summary == "" {
			t.Errorf("command %v has empty Summary", cmd.Path)
		}
		if cmd.Run == nil {
			t.Errorf("command %v has nil Run", cmd.Path)
		}
	}
}

