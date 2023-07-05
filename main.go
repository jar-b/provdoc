package main

import (
	"flag"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

// options stores flag values for use by downstream model operations
type options struct {
	schemafile string
}

func main() {
	var schemafile string
	flag.StringVar(&schemafile, "schemafile", "", "JSON file storing provider schema data")
	flag.Parse()

	m, err := newModel(options{schemafile: schemafile})
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		log.Fatal(err)
	}
}
