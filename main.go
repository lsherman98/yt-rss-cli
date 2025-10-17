package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lsherman98/yt-rss-cli/api"
	"github.com/lsherman98/yt-rss-cli/updater"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type viewState int

const (
	viewSetAPIKey viewState = iota
	viewMainMenu
	viewSelectPodcast
	viewEnterURL
	viewItemsTable
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)
)

type apiKeyCheckedMsg struct {
	hasKey bool
}

type podcastsLoadedMsg struct {
	podcasts []api.Podcast
	err      error
}

type urlAddedMsg struct {
	item api.Item
	err  error
}

type itemsLoadedMsg struct {
	items []api.Item
	err   error
}

type usageLoadedMsg struct {
	usage *api.UsageResponse
	err   error
}

type tickMsg time.Time

type menuItem string

func (i menuItem) FilterValue() string { return string(i) }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(menuItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	if index == m.Index() {
		fmt.Fprint(w, lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Render("> "+str))
	} else {
		fmt.Fprint(w, "  "+str)
	}
}

type model struct {
	state           viewState
	hasAPIKey       bool
	apiKeyInput     textinput.Model
	urlInput        textinput.Model
	mainMenu        list.Model
	podcastTable    table.Model
	itemsTable      table.Model
	podcasts        []api.Podcast
	selectedPodcast *api.Podcast
	items           []api.Item
	spinner         spinner.Model
	progressBar     progress.Model
	usage           *api.UsageResponse
	error           string
	message         string
	width           int
	height          int
	polling         bool
}

func initialModel() model {
	apiKeyInput := textinput.New()
	apiKeyInput.Placeholder = "Enter your API key"
	apiKeyInput.Focus()
	apiKeyInput.CharLimit = 256
	apiKeyInput.Width = 50

	urlInput := textinput.New()
	urlInput.Placeholder = "Paste YouTube URL here"
	urlInput.CharLimit = 500
	urlInput.Width = 80

	items := []list.Item{
		menuItem("Add YouTube URL"),
		menuItem("Set API Key"),
	}
	mainMenu := list.New(items, itemDelegate{}, 30, 8)
	mainMenu.Title = "Main Menu"
	mainMenu.SetShowStatusBar(false)
	mainMenu.SetFilteringEnabled(false)
	mainMenu.SetShowHelp(false)
	mainMenu.Styles.Title = titleStyle

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 40

	return model{
		state:       viewSetAPIKey,
		apiKeyInput: apiKeyInput,
		urlInput:    urlInput,
		mainMenu:    mainMenu,
		spinner:     s,
		progressBar: prog,
	}
}

func main() {
	updated, err := updater.CheckAndUpdate(version)
	if err != nil {
		fmt.Printf("⚠️  Update check failed: %v\n", err)
		fmt.Println("Continuing with current version...")
	}
	if updated {
		os.Exit(0)
	}

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}

func checkAPIKey() tea.Msg {
	_, err := api.GetApiKey()
	return apiKeyCheckedMsg{hasKey: err == nil}
}

func loadPodcasts() tea.Msg {
	podcasts, err := api.ListPodcasts()
	return podcastsLoadedMsg{podcasts: podcasts, err: err}
}

func addURL(podcastID, url string) tea.Cmd {
	return func() tea.Msg {
		item, err := api.AddUrlToPodcast(podcastID, url)
		return urlAddedMsg{item: item, err: err}
	}
}

func loadItems(podcastID string) tea.Cmd {
	return func() tea.Msg {
		items, err := api.GetPodcastItems(podcastID)
		return itemsLoadedMsg{items: items, err: err}
	}
}

func loadUsage() tea.Cmd {
	return func() tea.Msg {
		usage, err := api.GetUsage()
		return usageLoadedMsg{usage: usage, err: err}
	}
}

func parseCreatedTime(created string) time.Time {
	if created == "" {
		return time.Time{}
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999Z",
		"2006-01-02 15:04:05Z",
		"2006-01-02 15:04:05.999Z07:00",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, created); err == nil {
			return t
		}
	}

	return time.Time{}
}

func tick() tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tea.Batch(checkAPIKey, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case apiKeyCheckedMsg:
		m.hasAPIKey = msg.hasKey
		if msg.hasKey {
			m.state = viewMainMenu
			return m, loadUsage()
		} else {
			m.state = viewSetAPIKey
			m.apiKeyInput.Focus()
		}

	case usageLoadedMsg:
		if msg.err == nil {
			m.usage = msg.usage
		}

	case podcastsLoadedMsg:
		if msg.err != nil {
			m.error = msg.err.Error()
			m.state = viewMainMenu
		} else {
			m.podcasts = msg.podcasts
			m.error = ""
			m.buildPodcastTable()
			m.state = viewSelectPodcast
		}

	case urlAddedMsg:
		if msg.err != nil {
			m.error = msg.err.Error()
		} else {
			m.error = ""
			m.state = viewItemsTable
			m.polling = true
			cmds = append(cmds, loadItems(m.selectedPodcast.ID))
		}

	case itemsLoadedMsg:
		if msg.err != nil {
			m.error = msg.err.Error()
			m.polling = false
		} else {
			m.items = msg.items
			m.buildItemsTable()

			hasCreated := false
			allSuccess := true
			for _, item := range m.items {
				if item.Status == "CREATED" {
					hasCreated = true
					allSuccess = false
					break
				}
				if item.Status != "SUCCESS" {
					allSuccess = false
				}
			}

			if hasCreated && m.polling {
				cmds = append(cmds, tick())
			} else {
				m.polling = false
				if allSuccess && len(m.items) > 0 {
					cmds = append(cmds, loadUsage())
				}
			}
		}

	case tickMsg:
		if m.polling && m.selectedPodcast != nil {
			cmds = append(cmds, loadItems(m.selectedPodcast.ID))
		}

	case tea.KeyMsg:
		switch m.state {
		case viewSetAPIKey:
			switch msg.String() {
			case "ctrl+c", "esc":
				if m.hasAPIKey {
					m.state = viewMainMenu
					return m, nil
				}
				return m, tea.Quit
			case "enter":
				if m.apiKeyInput.Value() != "" {
					err := api.SetApiKey(m.apiKeyInput.Value())
					if err != nil {
						m.error = err.Error()
					} else {
						m.hasAPIKey = true
						m.message = "API key saved successfully!"
						m.apiKeyInput.SetValue("")
						m.state = viewMainMenu
						return m, loadUsage()
					}
				}
				return m, nil
			}

		case viewMainMenu:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter":
				selected := m.mainMenu.SelectedItem()
				if selected != nil {
					switch selected.(menuItem) {
					case "Set API Key":
						m.state = viewSetAPIKey
						m.apiKeyInput.Focus()
						m.error = ""
						m.message = ""
					case "Add YouTube URL":
						m.state = viewSelectPodcast
						m.error = ""
						m.message = ""
						return m, loadPodcasts
					}
				}
			}

		case viewSelectPodcast:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				m.state = viewMainMenu
				return m, nil
			case "enter":
				if m.podcastTable.Cursor() < len(m.podcasts) {
					m.selectedPodcast = &m.podcasts[m.podcastTable.Cursor()]
					m.state = viewEnterURL
					m.urlInput.Focus()
					m.urlInput.SetValue("")
					return m, nil
				}
			}

		case viewEnterURL:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				m.state = viewSelectPodcast
				m.urlInput.Blur()
				return m, nil
			case "enter":
				if m.urlInput.Value() != "" && m.selectedPodcast != nil {
					url := m.urlInput.Value()
					m.urlInput.SetValue("")
					return m, addURL(m.selectedPodcast.ID, url)
				}
			}

		case viewItemsTable:
			switch msg.String() {
			case "ctrl+c", "q":
				m.polling = false
				return m, tea.Quit
			case "a":
				m.state = viewEnterURL
				m.urlInput.Focus()
				m.urlInput.SetValue("")
				m.polling = false
				return m, nil
			case "m":
				m.state = viewMainMenu
				m.polling = false
				m.selectedPodcast = nil
				return m, loadUsage()
			}
		}
	}

	switch m.state {
	case viewSetAPIKey:
		m.apiKeyInput, cmd = m.apiKeyInput.Update(msg)
		cmds = append(cmds, cmd)
	case viewMainMenu:
		m.mainMenu, cmd = m.mainMenu.Update(msg)
		cmds = append(cmds, cmd)
	case viewSelectPodcast:
		m.podcastTable, cmd = m.podcastTable.Update(msg)
		cmds = append(cmds, cmd)
	case viewEnterURL:
		m.urlInput, cmd = m.urlInput.Update(msg)
		cmds = append(cmds, cmd)
	case viewItemsTable:
		m.itemsTable, cmd = m.itemsTable.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	if m.state == viewItemsTable && len(m.items) > 0 {
		m.buildItemsTable()
	}

	return m, tea.Batch(cmds...)
}

func (m *model) buildPodcastTable() {
	columns := []table.Column{
		{Title: "Title", Width: 60},
	}

	rows := []table.Row{}
	for _, p := range m.podcasts {
		rows = append(rows, table.Row{p.Title})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(false)

	t.SetStyles(s)
	m.podcastTable = t
}

func (m *model) buildItemsTable() {
	columns := []table.Column{
		{Title: "Title", Width: 60},
		{Title: "Status", Width: 20},
		{Title: "Created", Width: 30},
	}

	sortedItems := make([]api.Item, len(m.items))
	copy(sortedItems, m.items)
	sort.Slice(sortedItems, func(i, j int) bool {
		timeI := parseCreatedTime(sortedItems[i].Created)
		timeJ := parseCreatedTime(sortedItems[j].Created)

		if timeI.IsZero() && timeJ.IsZero() {
			return false
		}
		if timeI.IsZero() {
			return false
		}
		if timeJ.IsZero() {
			return true
		}

		return timeI.After(timeJ)
	})

	rows := []table.Row{}
	for _, item := range sortedItems {
		status := item.Status
		switch item.Status {
		case "CREATED":
			status = m.spinner.View() + " PROCESSING"
		case "ERROR":
			status = "❌ ERROR"
		case "SUCCESS":
			status = "✓ SUCCESS"
		}

		title := item.Title
		if title == "" {
			if item.Status == "CREATED" {
				title = "Processing..."
			} else {
				title = "(No title)"
			}
		}

		created := item.Created
		if created != "" {
			t := parseCreatedTime(created)
			if !t.IsZero() {
				created = t.Local().Format("Jan 2, 2006 3:04 PM")
			}
		} else {
			created = "-"
		}

		rows = append(rows, table.Row{title, status, created})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(len(rows)+2, 20)),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		BorderBottom(true).
		Bold(true)

	t.SetStyles(s)
	m.itemsTable = t
}

func formatBytes(bytes int) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	if bytes >= GB {
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	} else if bytes >= MB {
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	} else if bytes >= KB {
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	}
	return fmt.Sprintf("%d B", bytes)
}

func (m model) View() string {
	var s strings.Builder

	switch m.state {
	case viewSetAPIKey:
		s.WriteString(titleStyle.Render("Set API Key"))
		s.WriteString("\n\n")
		s.WriteString(m.apiKeyInput.View())
		s.WriteString("\n\n")
		if m.error != "" {
			s.WriteString(errorStyle.Render("Error: " + m.error))
			s.WriteString("\n")
		}
		s.WriteString(helpStyle.Render("Press Enter to save • Esc to cancel"))

	case viewMainMenu:
		if m.message != "" {
			s.WriteString(successStyle.Render(m.message))
			s.WriteString("\n\n")
		}
		s.WriteString(m.mainMenu.View())
		s.WriteString("\n")

		if m.usage != nil {
			s.WriteString("\n")
			usagePercent := 0.0
			if m.usage.Limit > 0 {
				usagePercent = float64(m.usage.Usage) / float64(m.usage.Limit)
			}
			usageText := fmt.Sprintf("Usage: %s / %s",
				formatBytes(m.usage.Usage),
				formatBytes(m.usage.Limit),
			)
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(usageText))
			s.WriteString("\n")
			s.WriteString(m.progressBar.ViewAs(usagePercent))
			s.WriteString("\n")
		}

		s.WriteString(helpStyle.Render("↑/↓: Navigate • Enter: Select • q: Quit"))

	case viewSelectPodcast:
		s.WriteString(titleStyle.Render("Select a Podcast"))
		s.WriteString("\n\n")
		if len(m.podcasts) == 0 {
			s.WriteString("No podcasts found.\n")
		} else {
			s.WriteString(m.podcastTable.View())
		}
		s.WriteString("\n")
		if m.error != "" {
			s.WriteString(errorStyle.Render("Error: " + m.error))
			s.WriteString("\n")
		}
		s.WriteString(helpStyle.Render("↑/↓: Navigate • Enter: Select • Esc: Back • q: Quit"))

	case viewEnterURL:
		s.WriteString(titleStyle.Render(fmt.Sprintf("Add URL to: %s", m.selectedPodcast.Title)))
		s.WriteString("\n\n")
		s.WriteString(m.urlInput.View())
		s.WriteString("\n\n")
		if m.error != "" {
			s.WriteString(errorStyle.Render("Error: " + m.error))
			s.WriteString("\n")
		}
		s.WriteString(helpStyle.Render("Press Enter to add URL • Esc: Back • q: Quit"))

	case viewItemsTable:
		s.WriteString(titleStyle.Render(fmt.Sprintf("Items for: %s", m.selectedPodcast.Title)))
		s.WriteString("\n\n")
		s.WriteString(m.itemsTable.View())
		s.WriteString("\n\n")
		if m.error != "" {
			s.WriteString(errorStyle.Render("Error: " + m.error))
			s.WriteString("\n")
		}
		if m.polling {
			s.WriteString(helpStyle.Render("Polling for updates... • a: Add another URL • m: Main menu • q: Quit"))
		} else {
			s.WriteString(helpStyle.Render("a: Add another URL • m: Main menu • q: Quit"))
		}
	}

	return s.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
