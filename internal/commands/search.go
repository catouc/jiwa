package commands

import (
	"context"
	"fmt"

	"github.com/andygrunwald/go-jira"
)

func (c *Command) Search(jqlQuery string) ([]jira.Issue, error) {
	issues, err := c.Client.Search(context.TODO(), jqlQuery)
	if err != nil {
		return nil, fmt.Errorf("could not search issues: %w", err)
	}

	return issues, nil
}
