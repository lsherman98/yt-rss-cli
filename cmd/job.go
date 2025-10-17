package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lsherman98/yt-rss-cli/utils"
	"github.com/spf13/cobra"
)

var createJobsCmd = &cobra.Command{
	Use:   "create [url]",
	Short: "Create a conversion job",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp := utils.CreateJobs(args[0])
		if resp == nil {
			fmt.Println("Error creating job")
			return
		}

		fmt.Println("Job created successfully!")
	},
}

var listJobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "List all jobs",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := ShowJobsList()
		if err != nil {
			fmt.Println("Error listing jobs:", err)
		}
	},
}

func ShowJobsList() (string, error) {
	jobs, err := utils.ListJobs()
	if err != nil {
		return "", fmt.Errorf("error listing jobs: %w", err)
	}

	if len(jobs) == 0 {
		fmt.Println("No jobs found.")
		return "", nil
	}

	items := make([]list.Item, len(jobs))
	for i, j := range jobs {
		items[i] = Item{title: fmt.Sprintf("%s (%s)", j.URL, j.Status), desc: j.ID}
	}

	listModel := NewList(items, "Jobs")
	if _, err := tea.NewProgram(listModel).Run(); err != nil {
		return "", fmt.Errorf("error running jobs viewer: %w", err)
	}

	return "", nil
}

func init() {
	rootCmd.AddCommand(createJobsCmd)
	rootCmd.AddCommand(listJobsCmd)
}
