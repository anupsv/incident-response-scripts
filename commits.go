package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// Commit represents a GitHub commit.
type Commit struct {
	SHA        string `json:"sha"`
	Repository string `json:"repository"`
	Commit     struct {
		Author struct {
			Name string `json:"name"`
			Date string `json:"date"`
		} `json:"author"`
		Message string `json:"message"`
	} `json:"commit"`
}

// FetchCommits fetches commits made by a specified user.
func FetchCommits(username, token, sortOrder string, dateRange int) ([]Commit, error) {
	var allCommits []Commit
	page := 1
	startDate := time.Now().AddDate(0, -dateRange, 0)

	for {
		var commits []Commit
		var err error
		commits, err = fetchCommitsPage(username, token, page)
		if err != nil {
			return nil, err
		}

		if len(commits) == 0 {
			break
		}

		for _, commit := range commits {
			commitDate, err := time.Parse(time.RFC3339, commit.Commit.Author.Date)
			if err != nil {
				fmt.Println("Error parsing commit date:", err)
				continue
			}

			if commitDate.After(startDate) {
				allCommits = append(allCommits, commit)
			}
		}
		page++
	}

	// Sort commits based on sortOrder
	sort.Slice(allCommits, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, allCommits[i].Commit.Author.Date)
		timeJ, _ := time.Parse(time.RFC3339, allCommits[j].Commit.Author.Date)
		if sortOrder == "asc" {
			return timeI.Before(timeJ)
		}
		return timeI.After(timeJ)
	})

	fmt.Printf("Total commits processed: %d\n", len(allCommits))

	return allCommits, nil
}

func fetchCommitsPage(username, token string, page int) ([]Commit, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/events?page=%d", username, page)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch commits: %s", resp.Status)
	}

	// Handle rate limit
	rateLimitRemaining, _ := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
	if rateLimitRemaining <= 0 {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var events []struct {
		Type    string `json:"type"`
		Payload struct {
			Commits []Commit `json:"commits"`
		} `json:"payload"`
	}

	err = json.Unmarshal(body, &events)
	if err != nil {
		return nil, err
	}

	var commits []Commit
	for _, event := range events {
		if event.Type == "PushEvent" {
			commits = append(commits, event.Payload.Commits...)
		}
	}

	return commits, nil
}
