package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func convertToLocalTime(dateStr string) string {
	// Parse the time in RFC3339 format (used by GitHub API)
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return "Invalid Date"
	}

	// Convert to local timezone
	localTime := t.Local()
	return localTime.Format("2006-01-02 15:04:05")
}

func printCommitsInTable(commits []Commit) {
	const maxMessageLength = 50
	header := "SHA\tDate\t\tAuthor\tMessage\tURL\tPR"
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))

	for _, commit := range commits {
		localDate := convertToLocalTime(commit.Commit.Author.Date)
		message := commit.Commit.Message
		url := fmt.Sprintf("https://github.com/%s/commit/%s", commit.Repository, commit.SHA)
		if len(message) > maxMessageLength {
			message = message[:maxMessageLength-3] + "..."
		}
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", commit.SHA, localDate, commit.Commit.Author.Name, commit.Commit.Message, url, commit.PRURL)

	}
}

func outputToConsole(commits []Commit) {
	printCommitsInTable(commits)
}

func outputToCSV(commits []Commit, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"SHA", "Date", "Author", "Message", "link"})

	fmt.Printf("Writing to CSV...\n")

	for _, commit := range commits {
		url := fmt.Sprintf("https://github.com/%s/commit/%s", commit.Repository, commit.SHA)
		writer.Write([]string{commit.SHA, convertToLocalTime(commit.Commit.Author.Date),
			commit.Commit.Author.Name, commit.Commit.Message, url})
	}

	return nil
}

func outputToJSON(commits []Commit, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(commits)
}
