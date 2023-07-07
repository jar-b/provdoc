package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var schemafile string

func main() {
	// slightly better usage output
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags]\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
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
