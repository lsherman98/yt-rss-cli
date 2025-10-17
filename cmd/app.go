package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lsherman98/yt-rss-cli/utils"
)

type modelState int

const (
	menuState modelState = iota
	authState
	pollingState
	postPollingMenuState
)

type actionResultMsg struct {
	name   string
	err    error
	itemId string
}

type pollMsg struct {
	status string
	err    error
}

type mainModel struct {
	state           modelState
	menu            menuModel
	auth            authModel
	spinner         spinner.Model
	postPollingMenu menuModel
	statusMessage   string
	actionInFlight  bool
	pollingItemId   string
	pollingStatus   string
	pollingErr      error
	pollingDone     bool
	pollingDoneMsg  string
}

func initialModel() mainModel {
	apiKey, _ := utils.GetApiKey()

	s := spinner.New()
	s.Spinner = spinner.Dot
	if apiKey == "" {
		return mainModel{
			state:         authState,
			menu:          initialMenuModel(),
			auth:          initialAuthModel(),
			spinner:       s,
			statusMessage: "No API key found. Please enter one to continue.",
		}
	}

	return mainModel{
		state:           menuState,
		menu:            initialMenuModel(),
		auth:            initialAuthModel(),
		spinner:         s,
		postPollingMenu: initialPostPollingMenu(),
		statusMessage:   "",
	}
}

func (m mainModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, tea.ClearScreen)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case actionResultMsg:
		m.actionInFlight = false
		if msg.err != nil {
			m.statusMessage = fmt.Sprintf("%s failed: %v", msg.name, msg.err)
		} else {
			m.statusMessage = fmt.Sprintf("%s completed", msg.name)
			if msg.itemId != "" {
				m.state = pollingState
				m.pollingItemId = msg.itemId
				m.statusMessage = ""
				return m, pollItem(m.pollingItemId)
			}
		}
		return m, nil
	case pollMsg:
		if msg.err != nil {
			m.pollingErr = msg.err
			m.pollingDone = true
			m.pollingDoneMsg = "Polling failed."
			m.state = postPollingMenuState
			return m, nil
		}

		m.pollingStatus = msg.status
		if msg.status == "SUCCESS" || msg.status == "ERROR" {
			m.pollingDone = true
			m.pollingDoneMsg = fmt.Sprintf("Polling finished with status: %s", msg.status)
			m.state = postPollingMenuState
			return m, nil
		}

		return m, pollItem(m.pollingItemId)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.state {
	case menuState:
		newMenu, newCmd := m.menu.Update(msg)
		m.menu = newMenu.(menuModel)
		cmd = newCmd

		if !m.actionInFlight {
			var actionCmd tea.Cmd

			switch m.menu.selected {
			case "Authenticate":
				m.state = authState
				m.menu.selected = ""
				m.statusMessage = ""
				return m, cmd
			case "Add URL to Podcast":
				m.actionInFlight = true
				m.statusMessage = "Launching add flow..."
				actionCmd = runAction("Add URL to Podcast", RunAddFlow)
			case "View Jobs":
				m.actionInFlight = true
				m.statusMessage = "Fetching jobs..."
				actionCmd = runAction("View Jobs", ShowJobsList)
			case "Open Downloads Folder":
				m.actionInFlight = true
				m.statusMessage = "Opening downloads folder..."
				actionCmd = runAction("Open Downloads Folder", OpenDownloadsFolder)
			case "View Usage":
				m.actionInFlight = true
				m.statusMessage = "Retrieving usage..."
				actionCmd = runAction("View Usage", PrintUsage)
			case "Exit":
				return m, tea.Quit
			}

			if actionCmd != nil {
				m.menu.selected = ""
				if cmd != nil {
					cmd = tea.Batch(cmd, actionCmd)
				} else {
					cmd = actionCmd
				}
			}
		}

	case pollingState:
		// Polling is handled by ticks, just wait for it to finish
		return m, nil
	case postPollingMenuState:
		newMenu, newCmd := m.postPollingMenu.Update(msg)
		m.postPollingMenu = newMenu.(menuModel)
		cmd = newCmd

		switch m.postPollingMenu.selected {
		case "Add another URL":
			m.state = menuState
			m.postPollingMenu.selected = ""
			m.statusMessage = "Launching add flow..."
			return m, runAction("Add URL to Podcast", RunAddFlow)
		case "Go to main menu":
			m.state = menuState
			m.postPollingMenu.selected = ""
			m.statusMessage = ""
		case "Exit":
			return m, tea.Quit
		}

	case authState:
		newAuth, newCmd := m.auth.Update(msg)
		m.auth = newAuth.(authModel)
		cmd = newCmd
		if m.auth.finished {
			m.state = menuState
			m.auth = initialAuthModel()
			m.statusMessage = "Authentication updated"
		}
	}

	return m, cmd
}

func (m mainModel) View() string {
	switch m.state {
	case authState:
		return m.auth.View()
	case pollingState:
		if m.pollingErr != nil {
			return fmt.Sprintf("Error polling: %v", m.pollingErr)
		}
		return fmt.Sprintf("%s Polling item %s... Status: %s", m.spinner.View(), m.pollingItemId, m.pollingStatus)
	case postPollingMenuState:
		return m.pollingDoneMsg + "\n\n" + m.postPollingMenu.View()
	default:
		if m.actionInFlight {
			return fmt.Sprintf("%s %s", m.spinner.View(), m.statusMessage)
		}

		view := m.menu.View()
		if m.statusMessage != "" {
			view += "\n" + m.statusMessage
		}
		return view
	}
}

func runAction(name string, fn func() (string, error)) tea.Cmd {
	return func() tea.Msg {
		itemId, err := fn()
		return actionResultMsg{name: name, err: err, itemId: itemId}
	}
}

func pollItem(itemId string) tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		item, err := utils.PollItem(itemId)
		if err != nil {
			return pollMsg{err: err}
		}
		return pollMsg{status: item.Status}
	})
}

func startApp() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
