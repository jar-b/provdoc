# provdoc
[![build](https://github.com/jar-b/provdoc/actions/workflows/build.yml/badge.svg)](https://github.com/jar-b/provdoc/actions/workflows/build.yml)

Terraform provider documentation in the terminal.

<img width="800" src="./demo/demo.gif" />

## Installation

```
go install github.com/jar-b/provdoc@latest
```

## Usage

```console
$ provdoc -h
Usage of provdoc:
  -schemafile string
        JSON file storing provider schema data
```

```sh
# Load live from the `terraform providers schema -json` command
provdoc

# Load from an exported JSON file
provdoc -schemafile schema.json
```

## Motivation

Writing Terraform can require frequent context switching between the editor
and [Terraform Registry](https://registry.terraform.io/), especially when
adopting a new, unfamiliar provider. `provdoc` utilizes the existing
documentation available from provider schemas to supply searchable documentation
directly in the terminal.

## Requirements

- [Terraform](https://www.terraform.io/)
- An initialized Terraform project OR exported JSON schema file.

## How does it work?

`provdoc` should be executed in a directory with an initialized Terraform project.
On startup, the program executes `terraform providers schema -json` (or reads in 
exported data if the `-schemafile` argument is provided), gathering up
the schema documentation for all providers currently configured in the project. Once
the schema is ingested, a text input enables searching the schema documentation by 
resource/data source name, and the resulting content is rendered into the viewport. 
As providers are added or removed, the program can be restarted and the additional 
schema documentation will be picked up automatically.

This project relies heavily on the following:

- [hashicorp/terraform-json](https://github.com/hashicorp/terraform-json) - Provider schema processing
- [hashicorp/terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) - Provider schema markdown rendering
- The [Charm](https://github.com/charmbracelet) ecosystem of command line tools, most notably:
  - [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
  - [charmbracelet/glamour](https://github.com/charmbracelet/glamour)

## Future enhancements

At this phase the project is mostly a proof of concept. Some initial ideas for
future enhancements include:

- [ ] Persistent display of loaded providers
- [ ] Fuzzy search for resource names
- [ ] Filtered search (resources versus data sources) 
- [ ] Alternate full screen display
- [ ] Example configuration generation

