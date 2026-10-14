package generator

import (
	"fmt"
	"io/fs"

	"crudding/internal/generator/common"
	"crudding/internal/manifest"
)

// CheckEndpointSupport valida o manifesto e confirma que a linguagem suporta
// "add endpoint", para a CLI falhar antes de pedir o nome ao usuário.
func CheckEndpointSupport(m manifest.Manifest) error {
	_, err := endpointLanguage(m)
	return err
}

// AddEndpoint cria um novo recurso CRUD no projeto em root, de acordo com a
// combinação registrada no project.crud, e o registra no roteamento existente.
func AddEndpoint(root string, m manifest.Manifest, name string, templates fs.FS) (common.Result, error) {
	language, err := endpointLanguage(m)
	if err != nil {
		return common.Result{}, err
	}
	patch, err := language.endpoint(root, m.Config(), name)
	if err != nil {
		return common.Result{}, err
	}
	return common.Apply(root, templates, patch)
}

func endpointLanguage(m manifest.Manifest) (Language, error) {
	language, ok := Lookup(m.Language)
	if !ok {
		return Language{}, fmt.Errorf("linguagem não suportada: %s", m.Language)
	}
	if err := language.validate(m.Config()); err != nil {
		return Language{}, fmt.Errorf("%s inválido: %w", manifest.FileName, err)
	}
	if language.endpoint == nil {
		return Language{}, fmt.Errorf("add endpoint ainda não suportado para %s", language.Label)
	}
	return language, nil
}
