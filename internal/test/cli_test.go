package test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crudding/internal/cli"
	"crudding/internal/manifest"
	"crudding/internal/project"
)

func chdirTemp(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prev)
	})
}

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
			t.Errorf("command %q has empty Summary", cmd.Args)
		}
		if cmd.Run == nil {
			t.Errorf("command %q has nil Run", cmd.Args)
		}
	}
}

func TestCLINewHelp(t *testing.T) {
	for _, arg := range []string{"-h", "--help", "help"} {
		var stdout, stderr bytes.Buffer
		code := cli.Run([]string{"new", arg}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("Run(new %q) exit code = %d, want 0", arg, code)
		}
		if stderr.Len() != 0 {
			t.Fatalf("Run(new %q) unexpected stderr = %q", arg, stderr.String())
		}
		got := stdout.String()
		for _, section := range []string{
			"como funciona:",
			"como cada flag funciona:",
			"name:<nome> ou n:<nome>",
			"exemplos:",
			"crudding new name:meu-app lang:go framework:gin arch:hexagonal",
			"crudding new n:meu-app l:go f:gin a:hexagonal",
			"lang/l:\n  python\n  go\n  node",
			"runtime/r:\n  js\n  ts",
			"framework/f:\n  raw",
			"arch/a:\n  flat",
		} {
			if !strings.Contains(got, section) {
				t.Fatalf("Run(new %q) stdout missing %q:\n%s", arg, section, got)
			}
		}
	}
}

func TestCLINewListOptionsByKey(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{
			args: []string{"new", "l"},
			want: "lang/l:\n  python\n  go\n  node\n",
		},
		{
			args: []string{"new", "lang"},
			want: "lang/l:\n  python\n  go\n  node\n",
		},
		{
			args: []string{"new", "r"},
			want: "runtime/r:\n  js\n  ts\n",
		},
		{
			args: []string{"new", "f"},
			want: "framework/f:\n  raw\n  fastapi\n",
		},
		{
			args: []string{"new", "l:go", "f"},
			want: "framework/f:\n  raw\n  gin\n  fiber\n  chi\n",
		},
		{
			args: []string{"new", "a"},
			want: "arch/a:\n  flat\n  layered\n  modular\n  hexagonal\n  clean_arch\n",
		},
	}

	for _, tc := range cases {
		t.Run(strings.Join(tc.args, "_"), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(tc.args, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("Run(%v) exit code = %d, stderr = %q", tc.args, code, stderr.String())
			}
			if !strings.Contains(stdout.String(), tc.want) {
				t.Fatalf("Run(%v) stdout = %q, want substring %q", tc.args, stdout.String(), tc.want)
			}
		})
	}
}

func TestCLINewNonInteractive(t *testing.T) {
	cases := []struct {
		name        string
		projectName string
		args        []string
	}{
		{
			name:        "full keys",
			projectName: "meu-app",
			args:        []string{"new", "name:meu-app", "lang:go", "framework:gin", "arch:hexagonal", "runtime:xpto"},
		},
		{
			name:        "short keys",
			projectName: "meu-app-curto",
			args:        []string{"new", "n:meu-app-curto", "l:go", "f:gin", "a:hexagonal", "r:xpto"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			chdirTemp(t, tmp)

			var stdout, stderr bytes.Buffer
			code := cli.Run(tc.args, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("Run(%v) exit code = %d, stderr = %q", tc.args, code, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("unexpected stderr = %q", stderr.String())
			}

			projectDir := filepath.Join(tmp, tc.projectName)
			_, m, err := manifest.Find(projectDir)
			if err != nil {
				t.Fatalf("manifest.Find() error = %v", err)
			}
			want := project.Config{
				ProjectName:  tc.projectName,
				Language:     "go",
				Framework:    "gin",
				Architecture: "hexagonal",
			}
			if m.Config() != want {
				t.Fatalf("manifest config = %+v, want %+v", m.Config(), want)
			}
			assertFilesExist(t, projectDir, "cmd/api/main.go", "internal/routes/routes.go")
		})
	}
}

func TestCLINewNonInteractiveNodeWithRuntime(t *testing.T) {
	tmp := t.TempDir()
	chdirTemp(t, tmp)

	var stdout, stderr bytes.Buffer
	args := []string{"new", "n:node-app", "l:node", "r:ts", "f:express", "a:layered"}
	code := cli.Run(args, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(%v) exit code = %d, stderr = %q", args, code, stderr.String())
	}

	projectDir := filepath.Join(tmp, "node-app")
	_, m, err := manifest.Find(projectDir)
	if err != nil {
		t.Fatalf("manifest.Find() error = %v", err)
	}
	want := project.Config{
		ProjectName:  "node-app",
		Language:     "node",
		Runtime:      "ts",
		Framework:    "express",
		Architecture: "layered",
	}
	if m.Config() != want {
		t.Fatalf("manifest config = %+v, want %+v", m.Config(), want)
	}
}

func TestCLINewNonInteractiveValidationErrors(t *testing.T) {
	tmp := t.TempDir()
	chdirTemp(t, tmp)

	cases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing flags",
			args: []string{"new", "name:meu-app", "lang:go"},
			want: "flags obrigatórias ausentes: framework/f, arch/a",
		},
		{
			name: "missing node runtime",
			args: []string{"new", "n:node-app", "l:node", "f:express", "a:flat"},
			want: "flags obrigatórias ausentes: runtime/r",
		},
		{
			name: "name key without value does not list options",
			args: []string{"new", "n"},
			want: "nenhum nome de projeto informado",
		},
		{
			name: "invalid project name",
			args: []string{"new", "name:../bad", "lang:go", "framework:raw", "arch:flat"},
			want: "nome inválido",
		},
		{
			name: "unknown flag",
			args: []string{"new", "n:meu-app", "l:go", "f:raw", "a:flat", "extra"},
			want: "flag desconhecida: extra",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(tc.args, &stdout, &stderr)
			if code != 1 {
				t.Fatalf("Run(%v) exit code = %d, want 1", tc.args, code)
			}
			if !strings.Contains(stderr.String(), tc.want) {
				t.Fatalf("stderr = %q, want substring %q", stderr.String(), tc.want)
			}
		})
	}
}

func TestCLIAddEndpointNonInteractive(t *testing.T) {
	tmp := t.TempDir()
	chdirTemp(t, tmp)

	var stdout, stderr bytes.Buffer
	newArgs := []string{"new", "n:meu-app", "l:go", "f:gin", "a:hexagonal"}
	if code := cli.Run(newArgs, &stdout, &stderr); code != 0 {
		t.Fatalf("Run(new) exit code = %d, stderr = %q", code, stderr.String())
	}

	projectDir := filepath.Join(tmp, "meu-app")
	chdirTemp(t, projectDir)

	stdout.Reset()
	stderr.Reset()
	addArgs := []string{"add", "endpoint", "product"}
	if code := cli.Run(addArgs, &stdout, &stderr); code != 0 {
		t.Fatalf("Run(%v) exit code = %d, stderr = %q", addArgs, code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Endpoint product criado.") {
		t.Fatalf("stdout = %q, missing endpoint summary", stdout.String())
	}
	assertFilesExist(t, projectDir, "internal/handlers/product.go", "internal/application/product.go")

	stdout.Reset()
	stderr.Reset()
	addFlagArgs := []string{"add", "endpoint", "n:order_item"}
	if code := cli.Run(addFlagArgs, &stdout, &stderr); code != 0 {
		t.Fatalf("Run(%v) exit code = %d, stderr = %q", addFlagArgs, code, stderr.String())
	}
	assertFilesExist(t, projectDir, "internal/handlers/order_item.go", "internal/application/order_item.go")

	stdout.Reset()
	stderr.Reset()
	badArgs := []string{"add", "endpoint", "Invalid-Name"}
	if code := cli.Run(badArgs, &stdout, &stderr); code != 1 {
		t.Fatalf("Run(%v) exit code = %d, want 1", badArgs, code)
	}
	if !strings.Contains(stderr.String(), "nome de endpoint inválido") {
		t.Fatalf("stderr = %q, want invalid endpoint error", stderr.String())
	}
}
