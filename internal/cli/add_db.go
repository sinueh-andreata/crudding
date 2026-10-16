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
)

func printDBSummary(stdout io.Writer, db string, lang string, result common.Result) {
	fmt.Fprintf(stdout, "\nBanco de dados %s adicionado com sucesso.\n", db)
	for _, file := range result.Created {
		fmt.Fprintf(stdout, "  criado     %s\n", file)
	}
	for _, file := range result.Modified {
		fmt.Fprintf(stdout, "  alterado   %s\n", file)
	}
	if lang == "go" {
		fmt.Fprintln(stdout, "\nPara atualizar as dependências Go, execute: go mod tidy")
	} else if lang == "python" {
		fmt.Fprintln(stdout, "\nPara instalar as dependências Python, execute: pip install -r requirements.txt")
	} else if lang == "node" {
		fmt.Fprintln(stdout, "\nPara instalar as dependências Node, execute: npm install")
	}
}

func parseDBArgs(args []string) (string, bool, error) {
	if len(args) == 0 {
		return "", false, nil
	}
	fs := flag.NewFlagSet("crudding add db", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var db string
	fs.StringVar(&db, "db", "", "banco de dados (postgres, mysql, mongo)")
	fs.StringVar(&db, "d", "", "banco de dados (postgres, mysql, mongo)")
	if err := fs.Parse(args); err != nil {
		return "", false, err
	}
	rest := fs.Args()
	if db != "" && len(rest) > 0 {
		return "", false, fmt.Errorf("argumentos inesperados: %s", strings.Join(rest, " "))
	}
	if len(rest) > 1 {
		return "", false, fmt.Errorf("argumentos inesperados: %s", strings.Join(rest[1:], " "))
	}
	if db == "" && len(rest) == 1 {
		token := rest[0]
		if key, val, ok := strings.Cut(token, ":"); ok && (key == "db" || key == "d") {
			db = val
		} else {
			db = token
		}
	}
	db = strings.TrimSpace(strings.ToLower(db))
	if err := generator.ValidateDatabase(db); err != nil {
		return "", false, err
	}
	return db, true, nil
}

func runAddDB(args []string, stdout, _ io.Writer) error {
	argDB, hasArg, err := parseDBArgs(args)
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

	db := argDB
	if !hasArg {
		selected, err := choose("escolha o banco de dados", generator.SupportedDatabases)
		if err != nil {
			return err
		}
		db = selected
	}

	result, err := generator.AddDatabase(root, m, db, crudding.Templates)
	if err != nil {
		return err
	}
	printDBSummary(stdout, db, m.Language, result)
	return nil
}

