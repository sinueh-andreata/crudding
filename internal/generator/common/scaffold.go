package common

import (
	"io/fs"
	"os"

	"crudding/internal/types"
)

// Plan descreve tudo o que um gerador de linguagem quer produzir. Os geradores
// apenas montam o Plan; criar o diretório, renderizar e limpar em caso de
// falha fica a cargo de Scaffold.
type Plan struct {
	Files []File
	Data  any
	// Prepare roda depois da criação do diretório raiz e antes da escrita dos
	// arquivos (ex.: criar pacotes Python com __init__.py). Opcional.
	Prepare func(root string) error
}

// Scaffold cria o diretório do projeto (nunca sobrescreve um existente),
// executa o Plan e remove tudo o que foi gerado se algum passo falhar.
func Scaffold(destination string, templates fs.FS, plan Plan) (err error) {
	root, err := CreateProjectDirectory(destination)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(root)
		}
	}()

	if plan.Prepare != nil {
		if err = plan.Prepare(root); err != nil {
			return err
		}
	}
	return WriteFiles(root, templates, plan.Data, plan.Files)
}

// BaseData reúne os campos de template comuns a todas as linguagens. Deve ser
// embutido nos templateData específicos para que {{.ProjectName}} etc.
// continuem acessíveis diretamente.
type BaseData struct {
	ProjectName  string
	Framework    string
	Architecture string
}

func NewBaseData(config types.ProjectConfig) BaseData {
	return BaseData{
		ProjectName:  config.ProjectName,
		Framework:    config.Framework,
		Architecture: config.Architecture,
	}
}

