package textInput

import (
	"fmt"
	"refiber/cmd/ui"
	"regexp"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	errMsg error
)

// A textnput.model contains the data for the textinput step.
// It has the required methods that make it a bubbletea.Model
type model struct {
	textInput textinput.Model
	err       error
	output    *string
	header    string
}

// sanitizeInput verifies that an input text string gets validated
func sanitizeInput(input string) error {
	matched, err := regexp.Match("^[a-zA-Z0-9_-]+$", []byte(input))
	if !matched {
		return fmt.Errorf("string violates the input regex pattern, err: %v", err)
	}
	return nil
}

// InitialTextInputModel initializes a textinput step
// with the given data
func InitialTextInputModel(output *string, header string) model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	ti.Validate = sanitizeInput

	return model{
		textInput: ti,
		err:       nil,
		output:    output,
		header:    ui.TextTitle.Render(header),
	}
}

// Init is called at the beginning of a textinput step
// and sets the cursor to blink
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// Update is called when "things happen", it checks for the users text input,
// and for Ctrl+C or Esc to close the program.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if len(m.textInput.Value()) > 1 {
				*m.output = m.textInput.Value()
				return m, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEsc:
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
	return fmt.Sprintf("%s\n\n%s\n\n",
		m.header,
		m.textInput.View(),
	)
}

func (m model) Err() string {
	return m.err.Error()
}
