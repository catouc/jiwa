package commands

import (
	"context"
	"github.com/andygrunwald/go-jira"
)

func (c *Command) IssueTypes(projectKey string) ([]jira.IssueType, error) {
	project, err := c.Client.GetProject(context.TODO(), projectKey)
	if err != nil {
		return nil, err
	}

	return project.IssueTypes, nil
}
