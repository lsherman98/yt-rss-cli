package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type promptModel struct {
	label  string
	text   textinput.Model
	cancel bool
}

func newPromptModel(label, placeholder string, width int) promptModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 256
	ti.Width = width
	ti.Focus()

	return promptModel{
		label: label,
		text:  ti,
	}
}

func (m promptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancel = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.text, cmd = m.text.Update(msg)
	return m, cmd
}

func (m promptModel) View() string {
	content := fmt.Sprintf("%s\n\n%s\n\n%s", m.label, m.text.View(), "(esc to cancel)")
	return content + "\n"
}

func promptForInput(label, placeholder string, width int) (string, bool, error) {
	model := newPromptModel(label, placeholder, width)
	p, err := tea.NewProgram(model).Run()
	if err != nil {
		return "", false, err
	}

	result := p.(promptModel)
	if result.cancel {
		return "", true, nil
	}

	return result.text.Value(), false, nil
}
