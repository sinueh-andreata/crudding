package cli

import (
	"fmt"
	"os"

	"crudding"
	"crudding/internal/generator"
	"crudding/internal/generator/common"
	"crudding/internal/manifest"
	"crudding/internal/project"
)

func printEndpointSummary(name string, result common.Result) {
	fmt.Printf("\nEndpoint %s criado.\n", name)
	for _, file := range result.Created {
		fmt.Printf("  criado     %s\n", file)
	}
	for _, file := range result.Modified {
		fmt.Printf("  alterado   %s\n", file)
	}
}

func runAddEndpoint() error {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}
	root, m, err := manifest.Find(workingDirectory)
	if err != nil {
		return err
	}
	fmt.Printf("Projeto %s: %s\n", root, m.Stack())
	if err := generator.CheckEndpointSupport(m); err != nil {
		return err
	}
	name, err := project.ReadEndpointName(os.Stdin, os.Stdout)
	if err != nil {
		return err
	}
	result, err := generator.AddEndpoint(root, m, name, crudding.Templates)
	if err != nil {
		return err
	}
	printEndpointSummary(name, result)
	return nil
}
