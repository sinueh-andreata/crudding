package common

import (
	"fmt"
)

// DatabaseFiles devolve a lista de arquivos de infraestrutura agnósticos de linguagem
// (docker-compose.yml, .env.example, migrations/) para o banco especificado.
func DatabaseFiles(db string) []File {
	if db == "" {
		return nil
	}
	var files []File
	files = append(files, File{
		TemplatePath: fmt.Sprintf("templates/databases/%s/docker-compose.yml.tmpl", db),
		OutputPath:   "docker-compose.yml",
	})
	files = append(files, File{
		TemplatePath: fmt.Sprintf("templates/databases/%s/.env.example.tmpl", db),
		OutputPath:   ".env.example",
	})
	switch db {
	case "postgres", "mysql":
		files = append(files,
			File{
				TemplatePath: fmt.Sprintf("templates/databases/%s/migrations/000001_init.up.sql.tmpl", db),
				OutputPath:   "migrations/000001_init.up.sql",
			},
			File{
				TemplatePath: fmt.Sprintf("templates/databases/%s/migrations/000001_init.down.sql.tmpl", db),
				OutputPath:   "migrations/000001_init.down.sql",
			},
		)
	case "mongo":
		files = append(files, File{
			TemplatePath: "templates/databases/mongo/migrations/000001_init.js.tmpl",
			OutputPath:   "migrations/000001_init.js",
		})
	}
	return files
}

