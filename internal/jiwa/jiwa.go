package jiwa

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	Username   string
	Password   string
	BaseURL    string
	APIVersion string
	HTTPClient *http.Client
}

func (c *Client) callAPI(ctx context.Context, method, endpoint string, params url.Values, body io.Reader) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/rest/api/%s/%s?%s", c.BaseURL, c.APIVersion, endpoint, params.Encode())
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("content-type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("failed to call API %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, nil
}

type CreateIssueInput struct {
	Project     string
	Summary     string
	Description string
	Labels      []string
	Component   string
	Assignee    string
	Type        string
}

// CreateIssue tries to create the issue in the target project
// if the creation was successful it returns the issue ID
func (c *Client) CreateIssue(ctx context.Context, input CreateIssueInput) (jira.Issue, error) {
	i := jira.Issue{
		Fields: &jira.IssueFields{
			Project:     jira.Project{Key: input.Project},
			Summary:     input.Summary,
			Description: input.Description,
			Type:        jira.IssueType{Name: input.Type},
			Labels:      input.Labels,
		},
	}

	bodyBytes, err := json.Marshal(i)
	if err != nil {
		return jira.Issue{}, fmt.Errorf("failed to marshal body: %w", err)
	}

	b, err := c.callAPI(ctx, http.MethodPost, "issue", nil, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return jira.Issue{}, fmt.Errorf("failed to create issue: %w", err)
	}

	var j jira.Issue
	err = json.Unmarshal(b, &j)
	if err != nil {
		return jira.Issue{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return j, nil
}

// GetIssue finds an issue based on its key
func (c *Client) GetIssue(ctx context.Context, key string) (jira.Issue, error) {
	b, err := c.callAPI(ctx, http.MethodGet, "issue/"+key, nil, nil)
	if err != nil {
		return jira.Issue{}, fmt.Errorf("failed to get issue: %w", err)
	}

	var j jira.Issue
	err = json.Unmarshal(b, &j)
	if err != nil {
		return jira.Issue{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return j, nil
}

func (c *Client) UpdateIssue(ctx context.Context, issue jira.Issue) error {
	body, err := json.Marshal(issue)
	if err != nil {
		return fmt.Errorf("failed to marshal input issue: %w", err)
	}

	_, err = c.callAPI(ctx, http.MethodPut, "issue/"+issue.Key, nil, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) AssignIssue(ctx context.Context, key string, assignee string) error {

func (c *Client) Search(ctx context.Context, jql string) ([]jira.Issue, error) {
	if jql == "" {
		return nil, errors.New("cannot search with empty search query")
	}

	params := url.Values{}
	params.Set("jql", jql)

	b, err := c.callAPI(ctx, http.MethodGet, "search", params, nil)
	if err != nil {
		return nil, err
	}

	searchResp := struct {
		StartAt    int          `json:"startAt"`
		MaxResults int          `json:"maxResults"`
		Total      int          `json:"total"`
		Issues     []jira.Issue `json:"issues"`
	}{}
	err = json.Unmarshal(b, &searchResp)
	if err != nil {
		return fmt.Errorf("failed to reassign ticket: %s", err)
	}

	return searchResp.Issues, nil
}
