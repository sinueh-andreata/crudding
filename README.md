# crudding

CLI interativo em Go 1.27 que gera projetos a partir de templates embutidos no executável.

## Executar

```sh
go run ./cmd/crudding
```

Use as setas para navegar, Espaço para selecionar e Enter para avançar. Escolha Python, Go ou Node; depois selecione o framework e uma arquitetura (`monorepo`, `camadas`, `clean arch`, `hexagonal` ou `4horse`). Para Node, escolha também JavaScript (`js`) ou TypeScript (`ts`). Antes de gerar, o crudding pede um nome; esse nome será usado como pasta dentro do diretório atual. Uma pasta existente nunca é sobrescrita. Pressione `q` ou Esc para sair.

Os arquivos dos templates são compilados dentro do executável, sem baixar templates durante a geração. Para instalar as dependências, siga o README criado no projeto: Python usa `python -m pip install -r requirements.txt`, Go usa `go mod tidy`, e Node usa `npm install`.

## Instalar o comando `crudding`

```sh
go install ./cmd/crudding
```

O Go instala o executável em `$(go env GOBIN)` ou, se `GOBIN` não estiver definido, em `$(go env GOPATH)/bin`. Esse diretório precisa estar no `PATH` para executar `crudding` diretamente.

Reinstale o executável após atualizar o crudding para incluir mudanças nos templates.

## Estrutura

```text
cmd/crudding/          ponto de entrada
internal/generator/  dispatcher da geração
internal/generator/common/ helpers compartilhados
internal/generator/python/ geração de projetos Python
internal/generator/golang/ geração de projetos Go
internal/generator/node/   geração de projetos Node
internal/models/     modelo do menu interativo
internal/project/    nome, validação e configuração de projetos
internal/types/      tipos compartilhados internamente
templates/           templates incorporados ao binário por assets.go
```