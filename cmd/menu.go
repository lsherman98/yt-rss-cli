package cmd

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuModel struct {
	list     list.Model
	selected string
}

type menuItem string

func (i menuItem) FilterValue() string { return "" }

type menuDelegate struct{}

func (d menuDelegate) Height() int                               { return 1 }
func (d menuDelegate) Spacing() int                              { return 0 }
func (d menuDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d menuDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(menuItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	var fn func(string) string
	if index == m.Index() {
		fn = func(s string) string {
			return lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")).Render("> " + s)
		}
	} else {
		fn = func(s string) string {
			return lipgloss.NewStyle().PaddingLeft(4).Render(s)
		}
	}

	fmt.Fprint(w, fn(str))
}

func initialMenuModel() menuModel {
	items := []list.Item{
		menuItem("Add URL to Podcast"),
		menuItem("Convert"),
		menuItem("Set API Key"),
		menuItem("Exit"),
	}

	l := list.New(items, menuDelegate{}, 25, 14)
	l.Title = "What would you like to do?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return menuModel{list: l}
}

func initialPostPollingMenu() menuModel {
	items := []list.Item{
		menuItem("Add another URL"),
		menuItem("Go to main menu"),
		menuItem("Exit"),
	}

	l := list.New(items, menuDelegate{}, 25, 14)
	l.Title = "What would you like to do?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return menuModel{list: l}
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(menuItem)
			if ok {
				m.selected = string(i)
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m menuModel) View() string {
	return "\n" + m.list.View()
}
