package selectInput

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/refiber/refiber-cli/cmd/ui"
)

// TODO: add required option to the model, if not required user can "ctrl+c" and "esc" to skip

type model struct {
	cursor  int
	output  *string
	choices []*string
	header  string
}

func InitialSelectInputModel(output *string, header string, options []*string) model {
	return model{
		output:  output,
		choices: options,
		header:  ui.TextTitle.Render(header),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			*m.output = *m.choices[m.cursor]
			return m, tea.Quit

		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	s := strings.Builder{}
	s.WriteString(m.header + "\n\n")

	for i := 0; i < len(m.choices); i++ {
		if m.cursor == i {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(*m.choices[i])
		s.WriteString("\n")
	}

	s.WriteString("\n")
	// s.WriteString("\n(press q to quit)\n")

	return s.String()
}
