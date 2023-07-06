package main

import (
	"flag"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

var schemafile string

func main() {
	flag.StringVar(&schemafile, "schemafile", "", "JSON file storing provider schema data")
	flag.Parse()

	m, err := newModel()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		log.Fatal(err)
	}
}
