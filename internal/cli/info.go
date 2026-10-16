package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"crudding/internal/manifest"
)

// InfoUsage gera o texto de ajuda do comando "crudding info".
func InfoUsage() string {
	var b strings.Builder
	b.WriteString("como funciona:\n")
	b.WriteString("  Lê o project.crud do projeto atual e exibe a stack configurada (linguagem,\n")
	b.WriteString("  framework e arquitetura) e os endpoints registrados.\n\n")
	b.WriteString("uso:\n")
	b.WriteString("  crudding info\n")
	return b.String()
}

func printInfoReport(stdout io.Writer, root string, m manifest.Manifest) {
	cfg := m.Config()
	fmt.Fprintf(stdout, "Projeto: %s (%s)\n", m.Name, root)
	fmt.Fprintf(stdout, "Stack: %s\n", m.Stack())
	fmt.Fprintln(stdout, formatLanguageLine(cfg))
	fmt.Fprintf(stdout, "Framework: %s\n", cfg.Framework)
	fmt.Fprintf(stdout, "Arquitetura: %s\n", cfg.Architecture)

	fmt.Fprintf(stdout, "\nEndpoints (%d):\n", len(m.Endpoints))
	if len(m.Endpoints) == 0 {
		fmt.Fprintln(stdout, "  (nenhum)")
	} else {
		for _, ep := range m.Endpoints {
			fmt.Fprintf(stdout, "  - %s\n", ep)
		}
	}
}

func runInfo(args []string, stdout, _ io.Writer) error {
	fs := flag.NewFlagSet("crudding info", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, InfoUsage())
			return nil
		}
		return err
	}

	tokens := fs.Args()
	if len(tokens) == 1 && tokens[0] == "help" {
		fmt.Fprint(stdout, InfoUsage())
		return nil
	}
	if len(tokens) > 0 {
		return fmt.Errorf("argumentos inesperados: %s", strings.Join(tokens, " "))
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}
	root, m, err := manifest.Find(workingDirectory)
	if err != nil {
		return err
	}
	printInfoReport(stdout, root, m)
	return nil
}
