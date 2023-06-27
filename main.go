package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	tfjson "github.com/hashicorp/terraform-json"
)

const width = 118

//go:embed templates/resource.tmpl
var resourceTemplate string

var (
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	viewportStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			PaddingRight(2).
			PaddingLeft(2)
)

func main() {
	m, err := newModel()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)

type model struct {
	err             error
	providerSchemas tfjson.ProviderSchemas
	renderer        *glamour.TermRenderer
	textarea        textarea.Model
	viewport        viewport.Model
}

func newModel() (*model, error) {
	ta := textarea.New()
	ta.Placeholder = "aws_instance"
	ta.Prompt = "┃ "
	ta.CharLimit = 200
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	ta.SetWidth(width)
	ta.SetHeight(1)
	ta.Focus()

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	vp := viewport.New(width, 25)
	vp.Style = viewportStyle

	rend, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}

	// Initialize provider schemas
	ps, err := readProviderSchemas()
	if err != nil {
		return nil, err
	}

	// TODO: display loaded provider names?
	// TODO: use resource list for auto-fill?
	var providers, resources []string
	for k, v := range ps.Schemas {
		providers = append(providers, k)
		for k2 := range v.ResourceSchemas {
			resources = append(resources, k2)
		}
	}

	vp.SetContent(fmt.Sprintf(`Loaded %d provider(s), %d resource(s).

Search results will be displayed here.`,
		len(providers), len(resources)))

	return &model{
		textarea:        ta,
		viewport:        vp,
		providerSchemas: ps,
		renderer:        rend,
		err:             nil,
	}, nil
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			var content string

			name := m.textarea.Value()
			match := m.SearchSchemas(name)
			if match == nil {
				content = fmt.Sprintf("No matches found for '%s'.\n", name)
			} else {
				raw, err := renderSchemaContent(name, match)
				if err != nil {
					// TODO: handle this
				}

				content, err = m.renderer.Render(raw)
				if err != nil {
					// TODO: handle this
				}
			}

			m.viewport.SetYOffset(0)
			m.viewport.SetContent(content)
			m.textarea.Reset()
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"Enter a resource name:\n\n%s\n\n%s\n%s",
		m.textarea.View(),
		m.viewport.View(),
		m.helpView(),
	)
}

func (m model) helpView() string {
	return helpStyle("  ↑/↓: Navigate • ctrl+c/esc: Quit\n")
}

func (m model) SearchSchemas(s string) *tfjson.Schema {
	// TODO: aggregate results in the case of multiple matches?
	// TODO: allow targeted searching between resources/data sources
	for _, prov := range m.providerSchemas.Schemas {
		if v, ok := prov.ResourceSchemas[s]; ok {
			return v
		}
		if v, ok := prov.DataSourceSchemas[s]; ok {
			return v
		}
	}
	return nil
}

func readProviderSchemas() (tfjson.ProviderSchemas, error) {
	var p tfjson.ProviderSchemas

	b, err := os.ReadFile("example-config/schema.json")
	if err != nil {
		return p, err
	}

	if err := p.UnmarshalJSON(b); err != nil {
		return p, err
	}

	return p, nil
}

type docData struct {
	Name     string
	Required []attribute
	Optional []attribute
	Computed []attribute
}

type attribute struct {
	Name        string
	Type        string
	Description string
}

func renderSchemaContent(name string, schema *tfjson.Schema) (string, error) {
	var req, opt, comp []attribute
	for k, v := range schema.Block.Attributes {
		if v.Required {
			req = append(req, attribute{Name: k, Type: v.AttributeType.GoString(), Description: v.Description})
		}
		if v.Optional {
			opt = append(opt, attribute{Name: k, Type: v.AttributeType.GoString(), Description: v.Description})
		}
		if v.Computed && !v.Optional {
			comp = append(comp, attribute{Name: k, Type: v.AttributeType.GoString(), Description: v.Description})
		}
	}
	data := docData{Name: name, Required: req, Optional: opt, Computed: comp}

	tmpl, err := template.New("resource").Parse(resourceTemplate)
	if err != nil {
		return "", err
	}

	b := &bytes.Buffer{}
	if err := tmpl.Execute(b, data); err != nil {
		return "", err
	}

	return b.String(), nil
}
