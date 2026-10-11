package main

import (
	"fmt"
	"os"
	"path/filepath"

	"crudding"
	"crudding/internal/generator"
	"crudding/internal/models"
	"crudding/internal/project"
	"crudding/internal/types"
)

func run() error {
	language, confirmed, err := models.ChooseOption("escolha a linguagem", []string{"python", "go", "node"})
	if err != nil || !confirmed {
		return err
	}

	runtime := ""
	if language == "node" {
		runtime, confirmed, err = models.ChooseOption("escolha o runtime", []string{"js", "ts"})
		if err != nil || !confirmed {
			return err
		}
	}

	framework, confirmed, err := models.ChooseOption("escolha o framework", frameworkOptionsFor(language))
	if err != nil || !confirmed {
		return err
	}
	architecture, confirmed, err := models.ChooseOption("escolha a arquitetura", architectureOptionsFor(language))
	if err != nil || !confirmed {
		return err
	}
	projectName, err := project.ReadProjectName(os.Stdin, os.Stdout)
	if err != nil {
		return err
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}
	config := types.ProjectConfig{
		ProjectName:  projectName,
		Language:     language,
		Runtime:      runtime,
		Framework:    framework,
		Architecture: architecture,
	}
	projectDirectory := filepath.Join(workingDirectory, projectName)
	if err := generator.Generate(projectDirectory, config, crudding.Templates); err != nil {
		return err
	}

	fmt.Printf("\nProjeto gerado em: %s\n", projectDirectory)
	if runtime != "" {
		fmt.Printf("Linguagem/motor: %s/%s\n", language, runtime)
	} else {
		fmt.Printf("Linguagem: %s\n", language)
	}
	fmt.Printf("Framework: %s\nArquitetura: %s\n", framework, architecture)
	return nil
}

func frameworkOptionsFor(language string) []string {
	switch language {
	case "python":
		return []string{"raw", "fastapi", "django", "flask"}
	case "go":
		return []string{"raw", "gin", "fiber", "chi"}
	case "node":
		return []string{"raw", "express", "nest", "fastify"}
	default:
		return nil
	}
}

func architectureOptionsFor(language string) []string {
	switch language {
	case "go":
		return []string{"flat", "layered", "modular", "hexagonal", "clean arch"}
	default:
		return []string{"monorepo", "camadas", "clean arch", "hexagonal", "4horse"}
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "crudding: %v\n", err)
		os.Exit(1)
	}
}