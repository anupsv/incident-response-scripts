package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// Commit represents a GitHub commit.
//type Commit struct {
//	SHA        string `json:"sha"`
//	Repository string `json:"repository"`
//	PRURL      string `json:"pr_url"`
//	Commit     struct {
//		Author struct {
//			Name string `json:"name"`
//			Date string `json:"date"`
//		} `json:"author"`
//		Message string `json:"message"`
//	} `json:"commit"`
//}

// FetchCommits fetches commits made by a specified user.
func FetchCommits(username, token, sortOrder string, dateRange int, mapToPR bool, githubAPIEndpoint string) ([]CustomCommitData, error) {
	var allCommits []CustomCommitData
	page := 1
	startDate := time.Now().AddDate(0, -dateRange, 0)

	for {
		var events []Event
		var err error
		events, err = fetchCommitsPage(username, token, page, githubAPIEndpoint)
		if err != nil {
			return nil, err
		}

		if len(events) == 0 {
			break
		}

		for _, event := range events {
			commitDate, err := time.Parse(time.RFC3339, event.CreatedAt)
			if err != nil {
				fmt.Println("Error parsing commit date:", err)
				continue
			}

			if commitDate.After(startDate) {

				for _, eachCommit := range event.Payload.Commits {
					allCommits = append(allCommits, CustomCommitData{
						Commit: eachCommit,
						Repo:   event.Repo,
						Date:   event.CreatedAt,
					})
				}
			}
		}

		if mapToPR {
			for i, eachCommit := range allCommits {
				prURL, err := findPRForCommit(eachCommit.Commit.SHA, eachCommit.Repo.Name, token, githubAPIEndpoint)
				if err != nil {
					fmt.Println("Error finding PR for commit:", err)
					continue
				}
				allCommits[i].PrUrl = prURL
			}
		}
		page++
	}

	// Sort commits based on sortOrder
	sort.Slice(allCommits, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, allCommits[i].Date)
		timeJ, _ := time.Parse(time.RFC3339, allCommits[j].Date)
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("unable to close body reader")
		}
	}(resp.Body)

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

	body, err := io.ReadAll(resp.Body)
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

func fetchCommitsPage(username, token string, page int, githubAPIEndpoint string) ([]Event, error) {
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("unable to close body reader")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch commits: %s", resp.Status)
	}

	// Handle rate limit
	rateLimitRemaining, _ := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
	if rateLimitRemaining <= 0 {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var receivedEvents []Event

	err = json.Unmarshal(body, &receivedEvents)
	if err != nil {
		return nil, err
	}

	var toSendEvents []Event
	for _, event := range receivedEvents {
		if event.Type == "PushEvent" {
			toSendEvents = append(toSendEvents, event)
		}
	}

	return toSendEvents, nil
}
