// Command genall gera todas as combinações suportadas de linguagem, runtime,
// framework e arquitetura, nomeando cada projeto como
// linguagem[-runtime]-framework-arquitetura (ex: go-gin-layered, node-ts-express-flat).
//
// Uso: go run ./cmd/genall [diretório de saída]   (padrão: ./generated)
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"crudding"
	"crudding/internal/generator"
	"crudding/internal/project"
)

func configs() []project.Config {
	var all []project.Config
	for _, language := range generator.Languages() {
		runtimes := language.Runtimes
		if len(runtimes) == 0 {
			runtimes = []string{""}
		}
		for _, runtime := range runtimes {
			for _, framework := range language.Frameworks {
				for _, architecture := range language.Architectures {
					// runtime vazio (python, go) não entra no nome
					parts := slices.DeleteFunc([]string{language.Name, runtime, framework, architecture}, func(part string) bool { return part == "" })
					all = append(all, project.Config{
						ProjectName:  strings.Join(parts, "-"),
						Language:     language.Name,
						Runtime:      runtime,
						Framework:    framework,
						Architecture: architecture,
					})
				}
			}
		}
	}
	return all
}

func run() error {
	output := "generated"
	if len(os.Args) > 1 {
		output = os.Args[1]
	}
	if err := os.MkdirAll(output, 0o755); err != nil {
		return fmt.Errorf("criando diretório de saída: %w", err)
	}

	var generated, skipped, failed int
	for _, config := range configs() {
		destination := filepath.Join(output, config.ProjectName)
		if _, err := os.Stat(destination); err == nil {
			fmt.Printf("pulado   %s (já existe)\n", config.ProjectName)
			skipped++
			continue
		}
		if err := generator.Generate(destination, config, crudding.Templates); err != nil {
			fmt.Fprintf(os.Stderr, "falhou   %s: %v\n", config.ProjectName, err)
			failed++
			continue
		}
		fmt.Printf("gerado   %s\n", config.ProjectName)
		generated++
	}

	fmt.Printf("\n%d gerados, %d pulados, %d falharam em %s\n", generated, skipped, failed, output)
	if failed > 0 {
		return fmt.Errorf("%d projetos falharam", failed)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "genall: %v\n", err)
		os.Exit(1)
	}
}
