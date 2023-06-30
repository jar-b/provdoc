package main

import (
	"flag"
	"log"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	tfjson "github.com/hashicorp/terraform-json"
)

var schemafile string

func main() {
	flag.StringVar(&schemafile, "schemafile", "", "JSON file storing provider schema data")
	flag.Parse()

	ps, err := loadProviderSchemas(schemafile)
	if err != nil {
		log.Fatal(err)
	}

	m, err := newModel(ps)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		log.Fatal(err)
	}
}

// loadProviderSchemas handles fetching configured provider schemas at
// startup. If a schemafile is specified, the schema is read from disk,
// otherwise, the schema is loaded on-demand by executing
// `terraform providers schema -json`.
func loadProviderSchemas(schemafile string) (tfjson.ProviderSchemas, error) {
	var (
		ps  tfjson.ProviderSchemas
		b   []byte
		err error
	)

	if schemafile == "" {
		b, err = exec.Command("terraform", "providers", "schema", "-json").Output()
	} else {
		b, err = os.ReadFile(schemafile)
	}
	if err != nil {
		return ps, err
	}

	if err := ps.UnmarshalJSON(b); err != nil {
		return ps, err
	}

	return ps, nil
}
