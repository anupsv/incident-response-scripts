package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func newFetchCommitsCmd(githubToken string, username string, dateRange int, sortOrder string, mapToPR bool,
	githubAPIEndpoint string, outputType, outputFile string, organization string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch-commits",
		Short: "Fetch commits",
		Long:  `Fetches commit data from a specified source.`,
		Run: func(cmd *cobra.Command, args []string) {

			validDateRange, err := validateDateRange(dateRange)
			if err != nil {
				fmt.Printf("Invalid date range: %s\n", err)
				return
			}

			commitEvents, err := FetchCommits(username, githubToken, sortOrder, validDateRange, mapToPR, githubAPIEndpoint)
			if err != nil {
				fmt.Println("Error fetching commits:", err)
				return
			}
			for _, commitEvent := range commitEvents {
				fmt.Printf("Commit: %s by %s on %s\n", commitEvent.Commit.SHA, commitEvent.Commit.Author.Name, commitEvent.Date)
			}

			switch outputType {
			case "console":
				outputToConsole(commitEvents)
			case "csv":
				if outputFile == "" {
					outputFile = "output.csv"
				}
				err := outputToCSV(commitEvents, outputFile)
				if err != nil {
					fmt.Println("Error writing CSV:", err)
				}
			case "json":
				if outputFile == "" {
					outputFile = "output.json"
				}
				err := outputToJSON(commitEvents, outputFile)
				if err != nil {
					fmt.Println("Error writing JSON:", err)
				}
			default:
				fmt.Println("Invalid output type. Use 'console', 'csv', or 'json'.")
			}

		},
	}

	// Adding flags for fetch-commits command
	cmd.Flags().StringVarP(&githubToken, "github-token", "t", "", "GitHub token")
	cmd.Flags().StringVarP(&username, "username", "u", "", "GitHub username")
	cmd.Flags().StringVarP(&sortOrder, "sort-order", "s", "desc", "Sort order of commits (asc or desc)")
	cmd.Flags().StringVarP(&organization, "organization", "o", "", "GitHub organization name")
	cmd.Flags().StringVarP(&outputType, "output-type", "x", "console", "Output type (console, csv, json)")
	cmd.Flags().StringVarP(&outputFile, "output-file", "f", "", "Output file name (optional)")
	cmd.Flags().IntVarP(&dateRange, "date-range", "d", 1, "Date range in months")
	cmd.Flags().BoolVarP(&mapToPR, "map-to-pr", "m", false, "Map commits to pull requests if applicable")
	cmd.Flags().StringVarP(&githubAPIEndpoint, "github-api-endpoint", "g", "https://api.github.com", "GitHub API endpoint")
	cmd.MarkFlagsOneRequired("github-token", "username")

	return cmd
}

func main() {
	var rootCmd = &cobra.Command{Use: "git-helper"}

	var githubToken, username string
	var dateRange int
	var sortOrder string
	var organization string
	var outputType, outputFile string
	var mapToPR bool
	var githubAPIEndpoint string

	var cmdFetchPRData = &cobra.Command{
		Use:   "fetch-pr-data",
		Short: "Fetch PR data",
		Long:  `Fetches data related to pull requests from a specified source.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Fetching PR data for user %s with token %s for the past %d months...\n", username, githubToken, dateRange)
			// Implement your logic here
		},
	}

	var cmdFetchCommits = newFetchCommitsCmd(githubToken, username, dateRange, sortOrder, mapToPR,
		githubAPIEndpoint, outputType, outputFile, organization)

	// Adding flags for fetch-pr-data command
	cmdFetchPRData.Flags().StringVarP(&githubToken, "github-token", "t", "", "GitHub token")
	cmdFetchPRData.Flags().StringVarP(&username, "username", "u", "", "GitHub username")
	cmdFetchPRData.Flags().IntVarP(&dateRange, "date-range", "d", 1, "Date range in months")
	cmdFetchPRData.Flags().StringVarP(&organization, "organization", "o", "", "GitHub organization name")
	cmdFetchPRData.Flags().StringVarP(&outputType, "output-type", "x", "console", "Output type (console, csv, json)")
	cmdFetchPRData.Flags().StringVarP(&outputFile, "output-file", "f", "", "Output file name (optional)")
	cmdFetchPRData.Flags().IntVarP(&dateRange, "date-range", "d", 1, "Date range in months")

	rootCmd.AddCommand(cmdFetchPRData)
	rootCmd.AddCommand(cmdFetchCommits)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func validateDateRange(dateRange int) (int, error) {
	if dateRange <= 0 {
		return 0, fmt.Errorf("date range must be a positive integer")
	}
	return dateRange, nil
}
