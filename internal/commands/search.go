package commands

import (
	"context"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"os"
)

func (c *Command) Search() ([]jira.Issue, error) {
	if len(os.Args) != 3 {
		fmt.Println("Usage: jiwa search \"<jql query>\"")
		os.Exit(1)
	}

	issues, err := c.Client.Search(context.TODO(), os.Args[2])
	if err != nil {
		return nil, fmt.Errorf("could not search issues: %w", err)
	}

	return issues, nil
}
