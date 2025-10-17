package cmd

import (
	"fmt"

	"github.com/lsherman98/yt-rss-cli/utils"
	"github.com/spf13/cobra"
)

var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Display API usage",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := PrintUsage()
		if err != nil {
			fmt.Println("Error getting usage:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(usageCmd)
}

func PrintUsage() (string, error) {
	usage, err := utils.GetUsage()
	if err != nil {
		return "", fmt.Errorf("error getting usage: %w", err)
	}

	fmt.Printf("Usage: %d / %d\n", usage.Used, usage.Usage)
	return "", nil
}
