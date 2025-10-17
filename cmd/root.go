package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ytrss",
	Short: "A CLI to interact with the yt-rss service",
	Long:  `A CLI to interact with the yt-rss service. Convert YouTube videos to audio and add them to your private podcast feeds.`,
	Run: func(cmd *cobra.Command, args []string) {
		startApp()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
