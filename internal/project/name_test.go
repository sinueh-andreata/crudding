package project

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadProjectName(t *testing.T) {
	var output bytes.Buffer
	name, err := ReadProjectName(strings.NewReader("Minha API\n"), &output)
	if err != nil || name != "Minha API" {
		t.Fatalf("ReadProjectName() = %q, %v", name, err)
	}
}

func TestProjectNameValidationAndSlug(t *testing.T) {
	for _, name := range []string{"", "../escape", "bad/name", "bad\\name", " name", "name "} {
		if _, err := ReadProjectName(strings.NewReader(name+"\n"), &bytes.Buffer{}); err == nil {
			t.Errorf("ReadProjectName(%q) should fail", name)
		}
	}
	if got := ProjectSlug("Minha API_2"); got != "minha-api-2" {
		t.Fatalf("ProjectSlug() = %q", got)
	}
}