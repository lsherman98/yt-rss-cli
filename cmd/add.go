package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lsherman98/yt-rss-cli/utils"
)

func RunAddFlow() (string, error) {
	usage, err := utils.GetUsage()
	if err != nil {
		return "", fmt.Errorf("error getting usage: %w", err)
	}
	fmt.Printf("Usage: %d / %d\n", usage.Used, usage.Usage)

	podcasts, err := utils.ListPodcasts()
	if err != nil {
		return "", fmt.Errorf("error listing podcasts: %w", err)
	}

	if len(podcasts) == 0 {
		fmt.Println("No podcasts found. Please create a podcast on the website first.")
		return "", nil
	}

	items := make([]list.Item, len(podcasts))
	for i, p := range podcasts {
		items[i] = Item{title: p.Name, desc: p.ID}
	}

	listModel := NewList(items, "Select a podcast to add the URL to")
	p, err := tea.NewProgram(listModel).Run()
	if err != nil {
		return "", fmt.Errorf("error running podcast selection: %w", err)
	}

	resultModel := p.(ListModel)
	if resultModel.Cancelled {
		fmt.Println("No podcast selected.")
		return "", nil
	}
	podcastID := resultModel.SelectedItem.desc

	url, cancelled, err := promptForInput("Enter the YouTube URL", "YouTube URL", 50)
	if err != nil {
		return "", fmt.Errorf("error prompting for URL: %w", err)
	}

	if cancelled || url == "" {
		fmt.Println("No URL entered.")
		return "", nil
	}

	item, err := utils.AddUrlToPodcast(podcastID, url)
	if err != nil {
		return "", fmt.Errorf("error adding URL to podcast: %w", err)
	}

	fmt.Println("Successfully added URL to podcast.")
	return item.ID, nil
}
