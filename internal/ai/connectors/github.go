package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// GitHubConnector collects evidence from GitHub repositories.
// Supports searching for commits, pull requests, and issues using GitHub's search API.
type GitHubConnector struct {
	config   Config
	client   *http.Client
	baseURL  string
	apiToken string
}

// NewGitHubConnector creates a new GitHub connector instance.
func NewGitHubConnector(cfg Config) (Connector, error) {
	baseURL := cfg.Endpoint
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}

	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &GitHubConnector{
		config:   cfg,
		client:   &http.Client{Timeout: timeout},
		baseURL:  baseURL,
		apiToken: cfg.APIKey,
	}, nil
}

// Name returns the connector identifier.
func (g *GitHubConnector) Name() string {
	return "github"
}

// Collect retrieves evidence from GitHub using the provided query.
// Query format: GitHub search syntax (e.g., "type:pr label:security", "is:issue author:username")
func (g *GitHubConnector) Collect(ctx context.Context, query string) ([]types.EvidenceEvent, error) {
	// Determine search type from query
	searchType := "code"
	if strings.Contains(query, "type:pr") {
		searchType = "pr"
		query = strings.ReplaceAll(query, "type:pr", "is:pr")
	} else if strings.Contains(query, "type:issue") {
		searchType = "issue"
		query = strings.ReplaceAll(query, "type:issue", "is:issue")
	} else if strings.Contains(query, "type:commit") {
		searchType = "commit"
		query = strings.ReplaceAll(query, "type:commit", "")
	}

	// Build API URL
	var endpoint string
	switch searchType {
	case "pr", "issue":
		endpoint = "/search/issues"
	case "commit":
		endpoint = "/search/commits"
	default:
		endpoint = "/search/code"
	}

	apiURL := fmt.Sprintf("%s%s?q=%s&per_page=100", g.baseURL, endpoint, url.QueryEscape(query))

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication and headers
	req.Header.Set("Authorization", "Bearer "+g.apiToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	// For commit searches, we need a different accept header
	if searchType == "commit" {
		req.Header.Set("Accept", "application/vnd.github.cloak-preview+json")
	}

	// Execute request
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for rate limiting
	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrRateLimited
	}

	// Check for auth errors
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrAuthFailed
	}

	// Check for other errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to evidence events
	events := make([]types.EvidenceEvent, 0, len(result.Items))
	for _, item := range result.Items {
		event, err := g.convertToEvent(searchType, item)
		if err != nil {
			// Log error but continue processing other items
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

// convertToEvent converts a GitHub API response item to an EvidenceEvent.
func (g *GitHubConnector) convertToEvent(searchType string, item json.RawMessage) (types.EvidenceEvent, error) {
	// Parse common fields
	var common struct {
		ID        int64     `json:"id"`
		HTMLURL   string    `json:"html_url"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		User      struct {
			Login string `json:"login"`
		} `json:"user"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
	}

	if err := json.Unmarshal(item, &common); err != nil {
		return types.EvidenceEvent{}, fmt.Errorf("failed to parse common fields: %w", err)
	}

	// Determine actor (commits use "author", others use "user")
	actor := common.User.Login
	if actor == "" {
		actor = common.Author.Login
	}

	// Create base event
	event := types.EvidenceEvent{
		ID:        fmt.Sprintf("github-%d", common.ID),
		Source:    "github",
		Type:      searchType,
		Timestamp: common.CreatedAt,
		Content:   "", // Will be populated below
		Metadata: map[string]interface{}{
			"html_url": common.HTMLURL,
			"actor":    actor,
			"raw":      json.RawMessage(item),
		},
	}

	// Add type-specific fields
	switch searchType {
	case "pr", "issue":
		var details struct {
			Title  string `json:"title"`
			Body   string `json:"body"`
			State  string `json:"state"`
			Labels []struct {
				Name string `json:"name"`
			} `json:"labels"`
		}
		if err := json.Unmarshal(item, &details); err == nil {
			event.Content = details.Title + "\n\n" + details.Body
			event.Metadata["title"] = details.Title
			event.Metadata["body"] = details.Body
			event.Metadata["state"] = details.State

			labels := make([]string, len(details.Labels))
			for i, label := range details.Labels {
				labels[i] = label.Name
			}
			event.Metadata["labels"] = labels
		}

	case "commit":
		var details struct {
			Commit struct {
				Message string `json:"message"`
				Author  struct {
					Name string `json:"name"`
				} `json:"author"`
			} `json:"commit"`
			SHA string `json:"sha"`
		}
		if err := json.Unmarshal(item, &details); err == nil {
			event.Content = details.Commit.Message
			event.Metadata["sha"] = details.SHA
			event.Metadata["author_name"] = details.Commit.Author.Name
		}
	}

	return event, nil
}

// Validate checks if the GitHub connector is properly configured.
func (g *GitHubConnector) Validate(ctx context.Context) error {
	if g.apiToken == "" {
		return fmt.Errorf("GitHub API token is required")
	}

	// Test connectivity with a simple API call
	req, err := http.NewRequestWithContext(ctx, "GET", g.baseURL+"/user", nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.apiToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("validation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrAuthFailed
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("validation failed with status %d", resp.StatusCode)
	}

	return nil
}
