# provdoc

Terraform provider documentation in the terminal.

<img width="800" src="./demo/demo.gif" />

## Installation

```
go install github.com/jar-b/provdoc
```

## Motivation

Writing Terraform can require frequent context switching between the editor
and [Terraform Registry](https://registry.terraform.io/), especially when
adopting a new, unfamiliar provider. `provdoc` utilizes the existing
documentation available from provider schemas to supply searchable documentation
directly in the terminal.

## Requirements

- [Terraform](https://www.terraform.io/)
- An initialized Terraform project

## How does it work?

`provdoc` should be executed in a directory with an initialized Terraform project.
On startup, the program executes `terraform providers schema -json` gathering up
the schema documentation for all providers currently configured in the project. Once
the schema is ingested, the [TUI](https://en.wikipedia.org/wiki/Text-based_user_interface)
(text-based user interface) program starts. This consists of a simple text input
and viewport, where resource/data source names can be entered into the text input and
the corresponding schema documentation is rendered into the viewport. As providers
are added or removed, the program can be restarted and the additional schema documentation
will be picked up automatically.

This project relies heavily on the following:

- [hashicorp/terraform-json](https://github.com/hashicorp/terraform-json) - Provider schema processing
- [hashicorp/terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) - Rendering schema documentation
- The [Charm](https://github.com/charmbracelet) ecosystem of command line tools, most notably:
  - [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
  - [charmbracelet/glamour](https://github.com/charmbracelet/glamour)

## Future enhancements

At this phase the project is mostly a proof of concept. Some initial ideas for
future enhancements include:

- [ ] Persistent display of loaded providers
- [ ] Support schema JSON file input
- [ ] Filtered search (resources versus data sources) 
- [ ] Fuzzy search for resource names
- [ ] Example configuration generation

