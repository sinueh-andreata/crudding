package common

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// LineRemoval descreve linhas a serem removidas de um arquivo existente.
type LineRemoval struct {
	Path  string   // relativo à raiz do projeto, com "/"
	Lines []string // linhas a serem removidas (correspondência sem espaços nas pontas)
}

// RemovalPlan descreve os arquivos, diretórios e linhas a serem removidos de
// um projeto existente.
type RemovalPlan struct {
	Files       []string      // caminhos relativos à raiz
	Directories []string      // diretórios a serem removidos se ficarem vazios
	Removals    []LineRemoval // alterações em arquivos existentes
}

// RemoveResult lista os caminhos (relativos à raiz) removidos e alterados por Remove.
type RemoveResult struct {
	Removed  []string
	Modified []string
}

// Remove executa um RemovalPlan: remove os arquivos indicados, limpa as linhas
// nos arquivos existentes e remove diretórios que ficarem vazios. Se a gravação
// dos arquivos modificados falhar, desfaz as alterações feitas nesses arquivos.
func Remove(root string, plan RemovalPlan) (result RemoveResult, err error) {
	originals := map[string][]byte{}
	edited := map[string][]byte{}
	var editOrder []string

	for _, removal := range plan.Removals {
		current, ok := edited[removal.Path]
		if !ok {
			target := filepath.Join(root, filepath.FromSlash(removal.Path))
			var readErr error
			current, readErr = os.ReadFile(target)
			if readErr != nil {
				return RemoveResult{}, fmt.Errorf("ler %s: %w", removal.Path, readErr)
			}
			originals[removal.Path] = current
			editOrder = append(editOrder, removal.Path)
		}
		newContent, modified := removeLines(current, removal.Lines)
		if modified {
			edited[removal.Path] = newContent
		}
	}

	defer func() {
		if err == nil {
			return
		}
		for _, rel := range editOrder {
			_ = os.WriteFile(filepath.Join(root, filepath.FromSlash(rel)), originals[rel], 0644)
		}
	}()

	for _, rel := range plan.Files {
		target := filepath.Join(root, filepath.FromSlash(rel))
		removeErr := os.Remove(target)
		if removeErr == nil {
			result.Removed = append(result.Removed, rel)
		} else if !errors.Is(removeErr, fs.ErrNotExist) {
			return RemoveResult{}, fmt.Errorf("remover arquivo %s: %w", rel, removeErr)
		}
	}

	for _, rel := range editOrder {
		newContent, ok := edited[rel]
		if !ok {
			continue
		}
		target := filepath.Join(root, filepath.FromSlash(rel))
		if writeErr := os.WriteFile(target, newContent, 0644); writeErr != nil {
			return RemoveResult{}, fmt.Errorf("gravar arquivo %s: %w", rel, writeErr)
		}
		result.Modified = append(result.Modified, rel)
	}

	dirs := slices.Clone(plan.Directories)
	slices.SortFunc(dirs, func(a, b string) int {
		depthA := strings.Count(a, "/")
		depthB := strings.Count(b, "/")
		if depthA != depthB {
			return cmp.Compare(depthB, depthA)
		}
		return cmp.Compare(len(b), len(a))
	})
	for _, dir := range dirs {
		target := filepath.Join(root, filepath.FromSlash(dir))
		_ = os.Remove(target) // Remove diretório se estiver vazio; ignora se não estiver ou não existir
	}

	return result, nil
}

func removeLines(content []byte, toRemove []string) ([]byte, bool) {
	if len(toRemove) == 0 {
		return content, false
	}
	removeSet := make(map[string]bool, len(toRemove))
	for _, l := range toRemove {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			removeSet[trimmed] = true
		}
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	modified := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if removeSet[trimmed] {
			modified = true
			continue
		}
		newLines = append(newLines, line)
	}

	if !modified {
		return content, false
	}
	return []byte(strings.Join(newLines, "\n")), true
}

