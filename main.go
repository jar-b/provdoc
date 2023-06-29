package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-docs/schemamd"
)

const (
	spacebar = " "

	viewportWidth  = 118
	viewportHeight = 25
)

var (
	cursorLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))
	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
	viewportStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			PaddingRight(2).
			PaddingLeft(2)

	// viewportKeyMap sets custom key bindings for the viewport.
	//
	// The default keybindings (j, k, u, d, etc.) for navigation can cause
	// the viewport to jump around during searches if not overridden.
	viewportKeyMap = viewport.KeyMap{
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", spacebar),
			key.WithHelp("pgdn", "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "½ page down"),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
	}
)

func main() {
	// Initialize provider schemas
	// TODO: optionally allow loading from file
	ps, err := loadProviderSchemas()
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

type (
	errMsg error
)

type model struct {
	err             error
	providerSchemas tfjson.ProviderSchemas
	renderer        *glamour.TermRenderer
	textinput       textinput.Model
	viewport        viewport.Model
}

func newModel(schemas tfjson.ProviderSchemas) (*model, error) {
	ti := textinput.New()
	ti.Placeholder = "aws_instance"
	ti.Prompt = "┃ "
	ti.CharLimit = 200
	ti.TextStyle = cursorLineStyle
	ti.PromptStyle = cursorStyle
	ti.Cursor.Style = cursorStyle
	ti.Focus()

	vp := viewport.New(viewportWidth, viewportHeight)
	vp.Style = viewportStyle
	vp.KeyMap = viewportKeyMap

	rend, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(viewportWidth),
	)
	if err != nil {
		return nil, err
	}

	// TODO: display loaded provider names?
	// TODO: use resource list for auto-fill?
	var providers, resources []string
	for k, v := range schemas.Schemas {
		providers = append(providers, k)
		for k2 := range v.ResourceSchemas {
			resources = append(resources, k2)
		}
	}

	vp.SetContent(fmt.Sprintf(`Loaded %d provider(s), %d resource(s).

Search results will be displayed here.`,
		len(providers), len(resources)))

	return &model{
		textinput:       ti,
		viewport:        vp,
		providerSchemas: schemas,
		renderer:        rend,
		err:             nil,
	}, nil
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textinput, tiCmd = m.textinput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			var content string

			name := m.textinput.Value()
			match := m.searchSchemas(name)
			if match == nil {
				content = fmt.Sprintf("No matches found for '%s'.\n", name)
			} else {
				b := &strings.Builder{}
				if err := schemamd.Render(match, b); err != nil {
					// TODO: handle this
				}

				var err error
				formatted := fmt.Sprintf("# %s\n\n%s\n\n%s", name, match.Block.Description, b.String())
				content, err = m.renderer.Render(formatted)
				if err != nil {
					// TODO: handle this
				}
			}

			m.viewport.SetYOffset(0)
			m.viewport.SetContent(content)
			m.textinput.Reset()
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
		m.textinput.View(),
		m.viewport.View(),
		m.helpView(),
	)
}

func (m model) helpView() string {
	return helpStyle.Render("  ↑/↓, PgUp/PgDn: Navigate • ctrl+c/esc: Quit\n")
}

func (m model) searchSchemas(s string) *tfjson.Schema {
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

func loadProviderSchemas() (tfjson.ProviderSchemas, error) {
	var p tfjson.ProviderSchemas

	b, err := exec.Command("terraform", "providers", "schema", "-json").Output()
	if err != nil {
		if b != nil {
			log.Printf(string(b))
		}
		return p, err
	}

	if err := p.UnmarshalJSON(b); err != nil {
		return p, err
	}

	return p, nil
}

func readProviderSchemas(filename string) (tfjson.ProviderSchemas, error) {
	var p tfjson.ProviderSchemas

	b, err := os.ReadFile(filename)
	if err != nil {
		return p, err
	}

	if err := p.UnmarshalJSON(b); err != nil {
		return p, err
	}

	return p, nil
}
