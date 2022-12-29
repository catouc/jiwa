package commands

import (
	"context"
	"github.com/andygrunwald/go-jira"
)

func (c *Command) Cat(issueID string) (jira.Issue, error) {
	issue, err := c.Client.GetIssue(context.TODO(), issueID)
	if err != nil {
		return jira.Issue{}, err
	}

	return issue, nil
}
