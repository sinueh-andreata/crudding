package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crudding/internal/manifest"
	"crudding/internal/project"
)

func TestWriteAndFindFromSubdirectory(t *testing.T) {
	root := t.TempDir()
	config := project.Config{ProjectName: "demo", Language: "python", Framework: "fastapi", Architecture: "modular"}
	if err := manifest.Write(root, manifest.New(config)); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "src", "app")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}

	found, m, err := manifest.Find(nested)
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if found != root {
		t.Errorf("Find() root = %q, want %q", found, root)
	}
	if m.Config() != config {
		t.Errorf("Config() = %+v, want %+v", m.Config(), config)
	}
	if got := m.Stack(); got != "python + fastapi + modular" {
		t.Errorf("Stack() = %q", got)
	}
}

func TestFindErrors(t *testing.T) {
	if _, _, err := manifest.Find(t.TempDir()); err == nil || !strings.Contains(err.Error(), manifest.FileName) {
		t.Errorf("Find() sem manifesto: err = %v", err)
	}

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, manifest.FileName), []byte(`{"version": 99}`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manifest.Find(root); err == nil || !strings.Contains(err.Error(), "versão") {
		t.Errorf("Find() versão inválida: err = %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, manifest.FileName), []byte(`not json`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manifest.Find(root); err == nil {
		t.Error("Find() JSON inválido deveria falhar")
	}
}
