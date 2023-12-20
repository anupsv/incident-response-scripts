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
	PRURL      string `json:"pr_url"`
	Commit     struct {
		Author struct {
			Name string `json:"name"`
			Date string `json:"date"`
		} `json:"author"`
		Message string `json:"message"`
	} `json:"commit"`
}

// FetchCommits fetches commits made by a specified user.
func FetchCommits(username, token, sortOrder string, dateRange int, mapToPR bool, githubAPIEndpoint string) ([]Commit, error) {
	var allCommits []Commit
	page := 1
	startDate := time.Now().AddDate(0, -dateRange, 0)

	for {
		var commits []Commit
		var err error
		commits, err = fetchCommitsPage(username, token, page, githubAPIEndpoint)
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

		if mapToPR {
			for i, commit := range allCommits {
				prURL, err := findPRForCommit(commit.SHA, commit.Repository, token, githubAPIEndpoint)
				if err != nil {
					fmt.Println("Error finding PR for commit:", err)
					continue
				}
				allCommits[i].PRURL = prURL
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

// findPRForCommit checks if the commit is associated with any pull requests.
func findPRForCommit(commitSHA, repo, token, githubAPIEndpoint string) (string, error) {
	apiUrl := fmt.Sprintf("%s/repos/%s/commits/%s/pulls", githubAPIEndpoint, repo, commitSHA)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.groot-preview+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == http.StatusForbidden {
		rateLimitReset, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Reset"))
		if err == nil {
			resetTime := time.Unix(int64(rateLimitReset), 0)
			return "", fmt.Errorf("rate limit exceeded, resets at %s", resetTime)
		}
		return "", fmt.Errorf("rate limit exceeded")
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch pull requests: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var pullRequests []struct {
		HTMLURL string `json:"html_url"`
	}

	err = json.Unmarshal(body, &pullRequests)
	if err != nil {
		return "", err
	}

	if len(pullRequests) > 0 {
		return pullRequests[0].HTMLURL, nil
	}

	return "", nil
}

func fetchCommitsPage(username, token string, page int, githubAPIEndpoint string) ([]Commit, error) {
	url := fmt.Sprintf("%s/users/%s/events?page=%d", githubAPIEndpoint, username, page)
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
