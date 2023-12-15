# provdoc
[![build](https://github.com/jar-b/provdoc/actions/workflows/build.yml/badge.svg)](https://github.com/jar-b/provdoc/actions/workflows/build.yml)

Terraform provider documentation in the terminal.

<img width="800" src="./demo/demo.gif" />

## Installation

```
go install github.com/jar-b/provdoc@latest
```

## Requirements

- [Terraform](https://www.terraform.io/)
- An initialized Terraform project OR exported JSON schema file.

## Usage

```console
$ provdoc -h
Usage: provdoc [flags]

Flags:
  -schemafile string
        JSON file storing provider schema data
```

```sh
# Load live from the `terraform providers schema -json` command
provdoc

# Load from an exported JSON file
provdoc -schemafile schema.json
```

`provdoc` should be executed in a directory with an initialized Terraform project.
On startup, the program executes `terraform providers schema -json` (or reads in
exported data if the `-schemafile` argument is provided), gathering up
the schema documentation for all providers currently configured in the project.
If providers are added or removed, the schema data can be reloaded with `Ctrl+R`.

Once the schema is loaded two search modes are available.

- `Schema` mode expects an exact resource or data source name as the search term, and
will render the resulting schema documentation to the viewport. Example search terms
are be [`random_string`](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/string)
or [`aws_instance`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/instance).

- `Resource` mode accepts any search term, and will list all resources or data sources
containing the term to the viewport. Example search terms are `random_` (ie. list all
resources in the random provider) or `aws_ec2`.

The active search mode is displayed in lower left corner, and can be toggled with
`Tab`/`Shift+Tab`.

## Motivation

Writing Terraform can require frequent context switching between the editor
and [Terraform Registry](https://registry.terraform.io/), especially when
adopting a new, unfamiliar provider. `provdoc` utilizes the existing
documentation available from provider schemas to supply searchable documentation
directly in the terminal.

> Documentation is parsed and rendered from the provider schema, so the utility
> of what is displayed is largely dependent on what the provider developer includes
> there. Some providers (for example the AWS Terraform Provider)
> maintain standalone registry documentation and leave the schema descriptions
> empty. In these cases the tool can still provide argument name references, but
> the content will be considerably less useful than the Terraform registry.

## Prior art

This project relies heavily on the following:

- [hashicorp/terraform-json](https://github.com/hashicorp/terraform-json) - Provider schema processing
- [hashicorp/terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) - Provider schema markdown rendering
- The [Charm](https://github.com/charmbracelet) ecosystem of command line tools, most notably:
  - [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
  - [charmbracelet/glamour](https://github.com/charmbracelet/glamour)

## Future enhancements

At this phase the project is mostly a proof of concept. Some initial ideas for
future enhancements include:

- [x] Display loaded providers at startup
- [x] Fuzzy search for resource names
- [x] Live reloading 
- [ ] Paged results (resources and data sources of the same name)
- [ ] Alternate full screen display
- [ ] Example configuration generation

