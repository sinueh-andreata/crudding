# crudding

Interactive CLI written in Go 1.27 that generates projects from templates embedded in the executable.

## Run

```sh
go run ./cmd/crudding
```

Use the arrow keys to navigate, Space to select, and Enter to continue. Choose Python, Go, or Node, then select a framework and architecture. Python architectures are `flat`, `layered`, `modular`, `hexagonal`, and `clean_arch`. Go architectures are `flat`, `layered`, `modular`, `hexagonal`, and `clean arch`. Node frameworks are `raw`, `express`, `fastify`, and `nest`; its architectures are `flat`, `layered`, `modular`, `hexagonal`, and `clean_arch`. Also choose JavaScript (`js`) or TypeScript (`ts`). Before generating the project, crudding prompts for a name and creates a directory with that name in the current working directory. Existing directories are never overwritten. Press `q` or Esc to exit.

Template files are compiled into the executable, so no templates are downloaded during project generation. To install dependencies, follow the README generated for the project: Python uses `python -m pip install -r requirements.txt`, Go uses `go mod tidy`, and Node uses `npm install`.

## Install the `crudding` Command

```sh
go install ./cmd/crudding
```

Go installs the executable in `$(go env GOBIN)` or, if `GOBIN` is not set, in `$(go env GOPATH)/bin`. This directory must be on your `PATH` to run `crudding` directly.

Reinstall the executable after updating crudding to include template changes.

## Project Structure

```text
cmd/crudding/                    application entry point
internal/generator/              generation dispatcher
internal/generator/common/       shared helpers
internal/generator/python/       Python project generation
internal/generator/golang/       Go project generation
internal/generator/node/         Node project generation
internal/models/                 interactive menu model
internal/project/                project naming, validation, and configuration
internal/types/                  shared internal types
templates/                       templates embedded into the binary by assets.go
```