package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
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
	modePadding     = 1
	viewportWidth   = 118
	viewportHeight  = 25
	viewportPadding = 2

	// wordwrapWidth adjusts relative to the viewport width to avoid
	// rendering into viewport padding
	wordwrapWidth = viewportWidth - 5
)

var (
	cursorLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))
	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
	inputHeadingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("212")).
				Bold(true)
	modeStyle = lipgloss.NewStyle().
			PaddingRight(modePadding).
			PaddingLeft(modePadding).
			Background(lipgloss.Color("62")).
			Bold(true)
	viewportStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			PaddingRight(viewportPadding)

	// viewportKeyMap sets custom key bindings for the viewport.
	//
	// The default keybindings (j, k, u, d, etc.) for navigation can cause
	// the viewport to jump around during searches if not overridden.
	viewportKeyMap = viewport.KeyMap{
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
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

type mode string

const (
	// loadingText is displayed in the viewport at startup and during reloads
	loadingText = "Loading providers..."

	// modeSchema expects searches for exact resource or data source names
	// and returns the corresponding schema.
	modeSchema mode = "Schema"
	// modeResource expects searches for partial resource or data source names
	// and returns a list of matching names from all configured providers.
	modeResource mode = "Resource"

	modeSchemaPlaceholder   = "aws_instance"
	modeResourcePlaceholder = "ec2"
)

// errMsg wraps an error in a tea.Msg to be handled by the model update method
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// schemaMsg is the tea.Msg structure returned when provider schemas are loaded
// from disk
type schemaMsg struct {
	ps tfjson.ProviderSchemas
}

type model struct {
	err             error
	searchMode      mode
	providerSchemas tfjson.ProviderSchemas
	renderer        *glamour.TermRenderer
	textinput       textinput.Model
	viewport        viewport.Model
}

func newModel() (*model, error) {
	rend, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(wordwrapWidth),
	)
	if err != nil {
		return nil, err
	}

	return &model{
		err:        nil,
		searchMode: modeSchema,
		renderer:   rend,
		textinput:  newTextInput(),
		viewport:   newViewport(),
	}, nil
}

func newTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = modeSchemaPlaceholder
	ti.Prompt = "➜ "
	ti.CharLimit = 200
	ti.TextStyle = cursorLineStyle
	ti.PromptStyle = cursorStyle
	ti.Cursor.Style = cursorStyle
	ti.Focus()
	return ti
}

func newViewport() viewport.Model {
	vp := viewport.New(viewportWidth, viewportHeight)
	vp.Style = viewportStyle
	vp.KeyMap = viewportKeyMap
	vp.SetContent(loadingText)
	return vp
}

func (m model) Init() tea.Cmd {
	return loadProviderSchemas
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textinput, tiCmd = m.textinput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case schemaMsg:
		m.providerSchemas = msg.ps
		m.viewport.SetContent(m.postLoadingViewportView())
		tiCmd = tea.Batch(tiCmd, textinput.Blink)

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyCtrlR:
			m.viewport.SetContent(loadingText)
			return m, loadProviderSchemas

		case tea.KeyTab, tea.KeyShiftTab:
			if m.searchMode == modeResource {
				m.searchMode = modeSchema
				m.textinput.Placeholder = modeSchemaPlaceholder
			} else {
				m.searchMode = modeResource
				m.textinput.Placeholder = modeResourcePlaceholder
			}

		case tea.KeyEnter:
			var content string
			var err error
			name := m.textinput.Value()

			if m.searchMode == modeSchema {
				content, err = m.searchSchemas(name)
			} else if m.searchMode == modeResource {
				content, err = m.searchResources(name)
			}
			if err != nil {
				m.err = err
				return m, tea.Quit
			}

			m.viewport.SetYOffset(0)
			m.viewport.SetContent(content)
			m.textinput.Reset()
		}
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nModel enountered an error: %v\n\n", m.err)
	}

	return fmt.Sprintf(`%s

%s

%s
%s %s
`,
		m.headingView(),
		m.textinput.View(),
		m.viewport.View(),
		m.modeView(),
		m.helpView(),
	)
}

func (m model) headingView() string {
	s := "Enter a resource name:"
	if m.searchMode == modeResource {
		s = "Enter a search term:"
	}
	return inputHeadingStyle.Render(s)
}

func (m model) helpView() string {
	return helpStyle.Render(" ↑/↓, PgUp/PgDn: Navigate • Tab: Toggle Mode • ctrl+c/esc: Quit")
}

func (m model) modeView() string {
	return modeStyle.Render(fmt.Sprintf("Mode: %s", m.searchMode))
}

// postLoadingViewportView will appear once provider data is loaded
func (m model) postLoadingViewportView() string {
	var names []string
	for k := range m.providerSchemas.Schemas {
		names = append(names, k)
	}
	sort.Strings(names)
	return fmt.Sprintf("Ready to search. %d providers detected:\n\n- %s",
		len(names), strings.Join(names, "\n- "))
}

// searchSchemas finds a resource or data source schema for
// the search term. Rendered markdown content is returned.
func (m model) searchSchemas(term string) (string, error) {
	// TODO: aggregate results in the case of multiple matches?
	// TODO: allow targeted searching between resources/data sources
	var match *tfjson.Schema
	for _, prov := range m.providerSchemas.Schemas {
		if v, ok := prov.ResourceSchemas[term]; ok {
			match = v
			break
		}
		if v, ok := prov.DataSourceSchemas[term]; ok {
			match = v
			break
		}
	}

	if match == nil {
		return notFoundContent(term), nil
	}
	b := &strings.Builder{}
	if err := schemamd.Render(match, b); err != nil {
		return "", err
	}
	formatted := fmt.Sprintf("# %s\n\n%s\n\n%s", term, match.Block.Description, b.String())

	return m.renderer.Render(formatted)
}

// searchSchemas finds all resources or data sources containing
// the search term. Rendered markdown content is returned.
func (m model) searchResources(term string) (string, error) {
	matches := indexProviderSchemasWithFilter(m.providerSchemas, term)
	if len(matches) == 0 {
		return notFoundContent(term), nil
	}

	// TODO: move parsing into a template
	b := &strings.Builder{}
	for _, match := range matches {
		b.WriteString(fmt.Sprintf("# %s\n\n", match.Name))
		if len(match.Resources) > 0 {
			b.WriteString("## Resources\n\n")
			for _, r := range match.Resources {
				b.WriteString(fmt.Sprintf("- `%s`\n", r))
			}
		}
		if len(match.DataSources) > 0 {
			b.WriteString("## Data Sources\n\n")
			for _, ds := range match.DataSources {
				b.WriteString(fmt.Sprintf("- `%s`\n", ds))
			}
		}
	}

	return m.renderer.Render(b.String())
}

// providerIndex stores resource and data source names for partial search
type providerIndex struct {
	Name        string
	Resources   []string
	DataSources []string
}

// indexProviderSchemasWithFilters returns an index of provider resources
// and data sources which contain the search term
func indexProviderSchemasWithFilter(ps tfjson.ProviderSchemas, term string) []providerIndex {
	var filteredIndex []providerIndex
	for k, v := range ps.Schemas {
		pi := providerIndex{Name: k}
		for r := range v.ResourceSchemas {
			if strings.Contains(r, term) {
				pi.Resources = append(pi.Resources, r)
			}
		}
		for ds := range v.DataSourceSchemas {
			if strings.Contains(ds, term) {
				pi.DataSources = append(pi.DataSources, ds)
			}
		}
		if len(pi.Resources) > 0 || len(pi.DataSources) > 0 {
			sort.Strings(pi.Resources)
			sort.Strings(pi.DataSources)
			filteredIndex = append(filteredIndex, pi)
		}
	}
	return filteredIndex
}

// loadProviderSchemas handles fetching configured provider schemas. If
// a schemafile is specified, the schema is read from disk, otherwise,
// the schema is loaded on-demand by executing `terraform providers schema -json`.
// The resource and data source names are also indexed for each provider.
func loadProviderSchemas() tea.Msg {
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
		return errMsg{err}
	}

	if err := ps.UnmarshalJSON(b); err != nil {
		return errMsg{err}
	}
	if len(ps.Schemas) == 0 {
		return errMsg{errors.New("no provider schemas found")}
	}

	return schemaMsg{ps}
}

func notFoundContent(term string) string {
	return fmt.Sprintf("No matches found for '%s'.\n", term)
}
