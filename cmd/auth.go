package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lsherman98/yt-rss-cli/utils"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with the yt-rss service",
	Run:   func(cmd *cobra.Command, args []string) {},
}

type (
	errMsg error
)

type authModel struct {
	textInput textinput.Model
	err       error
	finished  bool
}

func initialAuthModel() authModel {
	ti := textinput.New()
	ti.Placeholder = "API Key"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return authModel{
		textInput: ti,
		err:       nil,
		finished:  false,
	}
}

func (m authModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m authModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.finished = true
			return m, nil
		case tea.KeyEnter:
			apiKey := m.textInput.Value()
			err := utils.SetApiKey(apiKey)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.finished = true
			return m, nil
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m authModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error setting API key: %s", m.err)
	}

	return fmt.Sprintf(
		"Please enter your API key:\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}

func init() {
	rootCmd.AddCommand(authCmd)
}
