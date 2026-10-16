package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"crudding"
	"crudding/internal/generator"
	"crudding/internal/generator/common"
	"crudding/internal/manifest"
	"crudding/internal/project"
)

func printEndpointSummary(stdout io.Writer, name string, result common.Result) {
	fmt.Fprintf(stdout, "\nEndpoint %s criado.\n", name)
	for _, file := range result.Created {
		fmt.Fprintf(stdout, "  criado     %s\n", file)
	}
	for _, file := range result.Modified {
		fmt.Fprintf(stdout, "  alterado   %s\n", file)
	}
}

func parseEndpointArgs(args []string) (string, bool, error) {
	if len(args) == 0 {
		return "", false, nil
	}
	fs := flag.NewFlagSet("crudding add endpoint", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var name string
	fs.StringVar(&name, "name", "", "nome do endpoint (snake_case)")
	if err := fs.Parse(args); err != nil {
		return "", false, err
	}
	rest := fs.Args()
	if name != "" && len(rest) > 0 {
		return "", false, fmt.Errorf("argumentos inesperados: %s", strings.Join(rest, " "))
	}
	if len(rest) > 1 {
		return "", false, fmt.Errorf("argumentos inesperados: %s", strings.Join(rest[1:], " "))
	}
	if name == "" && len(rest) == 1 {
		token := rest[0]
		if key, val, ok := strings.Cut(token, ":"); ok && (key == "name" || key == "n") {
			name = val
		} else {
			name = token
		}
	}
	name = strings.TrimSpace(name)
	if err := project.ValidateEndpointName(name); err != nil {
		return "", false, err
	}
	return name, true, nil
}

func runAddEndpoint(args []string, stdout, _ io.Writer) error {
	argName, hasArg, err := parseEndpointArgs(args)
	if err != nil {
		return err
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}
	root, m, err := manifest.Find(workingDirectory)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Projeto %s: %s\n", root, m.Stack())
	if err := generator.CheckEndpointSupport(m); err != nil {
		return err
	}

	name := argName
	if !hasArg {
		name, err = project.ReadEndpointName(os.Stdin, stdout)
		if err != nil {
			return err
		}
	}

	result, err := generator.AddEndpoint(root, m, name, crudding.Templates)
	if err != nil {
		return err
	}
	printEndpointSummary(stdout, name, result)
	return nil
}
