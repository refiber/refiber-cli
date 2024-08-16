package textInput

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/refiber/refiber-cli/cmd/ui"
)

type (
	errMsg error
)

// TODO: add required option to the model, if not required user can "ctrl+c" and "esc" to skip
// TODO: add custom validation

// A textnput.model contains the data for the textinput step.
// It has the required methods that make it a bubbletea.Model
type model struct {
	textInput    textinput.Model
	err          error
	output       *string
	header       *string
	disabledExit bool
	validation   textinput.ValidateFunc
}

type Config struct {
	Header       string
	Placeholder  string
	Validation   textinput.ValidateFunc
	MaxChar      int
	Width        int
	DisabledExit bool
}

// InitialTextInputModel initializes a textinput step
// with the given data
func InitialTextInputModel(output *string, config *Config) model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 300
	ti.Width = 20

	m := model{
		err:       nil,
		output:    output,
		textInput: ti,
	}

	if config != nil {
		if config.Placeholder != "" {
			ti.Placeholder = config.Placeholder
		}

		if config.Header != "" {
			m.header = &config.Header
		}

		if config.Validation != nil {
			m.validation = config.Validation
		}

		if config.MaxChar > 0 {
			ti.CharLimit = config.MaxChar
		}

		if config.Width > 0 {
			ti.Width = config.Width
		}

		if config.DisabledExit {
			m.disabledExit = true
		}
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.validation != nil && m.textInput.Value() != "" {
		err := m.validation(m.textInput.Value())
		if err != nil {
			m.err = err
		} else {
			m.err = nil
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.textInput.Value() == "" || m.err != nil {
				break
			}

			*m.output = m.textInput.Value()
			m.textInput.Blur()
			return m, tea.Quit
		case tea.KeyCtrlC:
			if m.disabledExit {
				return m, cmd
			}

			m.textInput.Blur()
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View is called to draw the textinput step
func (m model) View() string {
	if m.header == nil {
		if m.err != nil {
			return fmt.Sprintf("%s\n\n%s\n\n", m.textInput.View(), ui.TextError.PaddingLeft(2).Render(m.err.Error()))
		}

		return fmt.Sprintf("%s\n\n", m.textInput.View())
	}

	if m.err != nil {
		return fmt.Sprintf("%s\n\n%s\n\n%s\n\n",
			*m.header,
			m.textInput.View(),
			ui.TextError.PaddingLeft(2).Render(m.err.Error()),
		)
	}

	return fmt.Sprintf("%s\n\n%s\n\n",
		*m.header,
		m.textInput.View(),
	)
}

func (m model) Err() string {
	return m.err.Error()
}
