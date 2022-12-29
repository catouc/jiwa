package commands

import (
	"context"
	"fmt"
	"github.com/andygrunwald/go-jira"
)

func (c *Command) Edit(issueID string) (string, error) {
	summary, description, err := GetIssueIntoEditor(c.Client, issueID)
	if err != nil {
		return "", fmt.Errorf("failed to get summary and description: %w", err)
	}

	err = c.Client.UpdateIssue(context.TODO(), jira.Issue{
		Key: issueID,
		Fields: &jira.IssueFields{
			Summary:     summary,
			Description: description,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to update issue: %w", err)
	}

	return issueID, nil
}
