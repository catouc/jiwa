package commands

import (
	"context"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"os"
)

func (c *Command) IssueTypes() ([]jira.IssueType, error) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: jiwa issue-type <project-key>")
		os.Exit(1)
	}

	project, err := c.Client.GetProject(context.TODO(), os.Args[2])
	if err != nil {
		return nil, err
	}

	return project.IssueTypes, nil
}
