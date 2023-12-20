package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{Use: "myapp"}

	var githubToken, username string
	var dateRange int
	var sortOrder string
	var organization string
	var outputType, outputFile string

	var cmdFetchPRData = &cobra.Command{
		Use:   "fetch-pr-data",
		Short: "Fetch PR data",
		Long:  `Fetches data related to pull requests from a specified source.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Fetching PR data for user %s with token %s for the past %d months...\n", username, githubToken, dateRange)
			// Implement your logic here
		},
	}

	var cmdFetchCommits = &cobra.Command{
		Use:   "fetch-commits",
		Short: "Fetch commits",
		Long:  `Fetches commit data from a specified source.`,
		Run: func(cmd *cobra.Command, args []string) {

			validDateRange, err := validateDateRange(dateRange)
			if err != nil {
				fmt.Printf("Invalid date range: %s\n", err)
				return
			}

			commits, err := FetchCommits(username, githubToken, sortOrder)
			if err != nil {
				fmt.Println("Error fetching commits:", err)
				return
			}
			for _, commit := range commits {
				fmt.Printf("Commit: %s by %s on %s\n", commit.SHA, commit.Commit.Author.Name, commit.Commit.Author.Date)
			}

			switch outputType {
			case "console":
				outputToConsole(commits)
			case "csv":
				if outputFile == "" {
					outputFile = "output.csv"
				}
				err := outputToCSV(commits, outputFile)
				if err != nil {
					fmt.Println("Error writing CSV:", err)
				}
			case "json":
				if outputFile == "" {
					outputFile = "output.json"
				}
				err := outputToJSON(commits, outputFile)
				if err != nil {
					fmt.Println("Error writing JSON:", err)
				}
			default:
				fmt.Println("Invalid output type. Use 'console', 'csv', or 'json'.")
			}

		},
	}

	// Adding flags for fetch-pr-data command
	cmdFetchPRData.Flags().StringVarP(&githubToken, "github-token", "t", "", "GitHub token")
	cmdFetchPRData.Flags().StringVarP(&username, "username", "u", "", "GitHub username")
	cmdFetchPRData.Flags().IntVarP(&dateRange, "date-range", "d", 1, "Date range in months")
	cmdFetchPRData.Flags().StringVarP(&organization, "organization", "o", "", "GitHub organization name")
	cmdFetchPRData.Flags().StringVarP(&outputType, "output-type", "x", "console", "Output type (console, csv, json)")
	cmdFetchPRData.Flags().StringVarP(&outputFile, "output-file", "f", "", "Output file name (optional)")
	cmdFetchPRData.Flags().IntVarP(&dateRange, "date-range", "d", 1, "Date range in months")

	// Adding flags for fetch-commits command
	cmdFetchCommits.Flags().StringVarP(&githubToken, "github-token", "t", "", "GitHub token")
	cmdFetchCommits.Flags().StringVarP(&username, "username", "u", "", "GitHub username")
	cmdFetchCommits.Flags().IntVarP(&dateRange, "date-range", "d", 1, "Date range in months")
	cmdFetchCommits.Flags().StringVarP(&sortOrder, "sort-order", "s", "desc", "Sort order of commits (asc or desc)")
	cmdFetchCommits.Flags().StringVarP(&organization, "organization", "o", "", "GitHub organization name")
	cmdFetchCommits.Flags().StringVarP(&outputType, "output-type", "x", "console", "Output type (console, csv, json)")
	cmdFetchPRData.Flags().StringVarP(&outputFile, "output-file", "f", "", "Output file name (optional)")
	cmdFetchCommits.Flags().IntVarP(&dateRange, "date-range", "d", 1, "Date range in months")

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
