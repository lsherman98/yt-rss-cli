package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/lsherman98/yt-rss-cli/utils"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download [job-id]",
	Short: "Download the audio file for a completed job",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jobID := args[0]

		fmt.Println("Downloading audio...")
		filepath, err := DownloadJob(jobID)
		if err != nil {
			fmt.Println("Error downloading file:", err)
			return
		}

		fmt.Println("Successfully downloaded audio to:", filepath)
	},
}

var openDownloadsCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the downloads directory",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := OpenDownloadsFolder()
		if err != nil {
			fmt.Println("Error opening downloads directory:", err)
		}
	},
}

func DownloadJob(jobID string) (string, error) {
	job, err := utils.GetJob(jobID)
	if err != nil {
		return "", fmt.Errorf("error checking job status: %w", err)
	}

	if job.Status != "SUCCESS" {
		return "", fmt.Errorf("job is not ready for download. status: %s", job.Status)
	}

	filepath, err := utils.DownloadFile(jobID)
	if err != nil {
		return "", fmt.Errorf("error downloading file: %w", err)
	}

	return filepath, nil
}

func OpenDownloadsFolder() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %w", err)
	}
	downloadsDir := dir + "/downloads"

	var openCmd string
	switch runtime.GOOS {
	case "linux":
		openCmd = "xdg-open"
	case "darwin":
		openCmd = "open"
	case "windows":
		openCmd = "explorer"
	default:
		return "", fmt.Errorf("unsupported operating system")
	}

	if err := exec.Command(openCmd, downloadsDir).Start(); err != nil {
		return "", fmt.Errorf("error executing open command: %w", err)
	}

	return "", nil
}

func RunDownloadFlow() error {
	jobID, cancelled, err := promptForInput("Enter the Job ID to download", "Job ID", 40)
	if err != nil {
		return fmt.Errorf("error prompting for job ID: %w", err)
	}

	if cancelled {
		fmt.Println("Download cancelled.")
		return nil
	}

	if jobID == "" {
		fmt.Println("No Job ID entered.")
		return nil
	}

	fmt.Println("Downloading audio...")
	filepath, err := DownloadJob(jobID)
	if err != nil {
		return err
	}

	fmt.Println("Successfully downloaded audio to:", filepath)
	return nil
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(openDownloadsCmd)
}
