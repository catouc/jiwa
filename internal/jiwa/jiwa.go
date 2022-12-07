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
	"strings"
)

type Client struct {
	Username   string
	Password   string
	Token      string
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

	switch {
	case c.Username != "" && c.Password != "":
		req.SetBasicAuth(c.Username, c.Password)
	case c.Token != "":
		req.Header.Set("Authorization", "Bearer "+c.Token)
	default:
		return nil, errors.New("either username+password need to be set or token")
	}
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
	i := jira.Issue{
		Key: key,
		Fields: &jira.IssueFields{
			Assignee: &jira.User{Name: assignee},
		},
	}

	return c.UpdateIssue(ctx, i)
}

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
		fmt.Println(string(b))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return searchResp.Issues, nil
}

func (c *Client) LabelIssue(ctx context.Context, key string, labels ...string) error {
	if len(labels) == 0 {
		return errors.New("need to supply at least one label")
	}

	i := jira.Issue{
		Key:    key,
		Fields: &jira.IssueFields{Labels: labels},
	}

	return c.UpdateIssue(ctx, i)
}

func (c *Client) ListIssueTransitions(ctx context.Context, key string) ([]jira.Transition, error) {
	b, err := c.callAPI(ctx, http.MethodGet, "issue/"+key+"/transitions", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list transitions: %w", err)
	}

	var resp struct {
		Transitions []jira.Transition `json:"transitions"`
	}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarhal response: %w", err)
	}

	return resp.Transitions, nil
}

type TransitionRequest struct {
	Transition Transition `json:"transition"`
}

type Transition struct {
	ID string `json:"id"`
}

func (c *Client) TransitionIssue(ctx context.Context, key string, status string) error {
	transitions, err := c.ListIssueTransitions(ctx, key)
	if err != nil {
		return fmt.Errorf("could not list transitions: %w", err)
	}

	status = strings.ToLower(status)

	validTransitions := make([]string, len(transitions), len(transitions))
	transitionID := ""
	for _, t := range transitions {
		if strings.ToLower(t.Name) == status {
			transitionID = t.ID
		}

		validTransitions = append(validTransitions, t.Name)
	}

	if transitionID == "" {
		return fmt.Errorf(
			"could not find %s as a valid transition for %s, valid transitions are: %s",
			status,
			key,
			strings.Join(validTransitions, ","),
		)
	}

	tr := jira.CreateTransitionPayload{
		Transition: jira.TransitionPayload{ID: transitionID},
	}
	body, err := json.Marshal(&tr)
	if err != nil {
		return fmt.Errorf("failed to marshal transition request: %w", err)
	}

	_, err = c.callAPI(ctx, http.MethodPost, "issue/"+key+"/transitions", nil, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to transition issue to %s: %w", status, err)
	}

	return nil
}

func (c *Client) GetProject(ctx context.Context, key string) (jira.Project, error) {
	b, err := c.callAPI(ctx, http.MethodGet, "project/"+key, nil, nil)
	if err != nil {
		return jira.Project{}, fmt.Errorf("failed to get project %s: %w", key, err)
	}

	var result jira.Project
	err = json.Unmarshal(b, &result)
	if err != nil {
		return jira.Project{}, fmt.Errorf("failed to unmarshal project response: %w", err)
	}

	return result, nil
}
