package commands

import (
	"context"
	"fmt"
	"github.com/andygrunwald/go-jira"
)

func (c *Command) List(userFlag, projectFlag, statusFlag string) ([]jira.Issue, error) {
	var user string
	switch userFlag {
	case "empty":
		user = "AND assignee is EMPTY"
	case "":
		user = ""
	default:
		user = "AND assignee= \"" + userFlag + "\""
	}

	project := c.Config.DefaultProject
	if projectFlag != "" {
		project = projectFlag
	}

	jql := fmt.Sprintf("project=%s AND status=\"%s\" %s", project, statusFlag, user)
	issues, err := c.Client.Search(context.TODO(), jql)
	if err != nil {
		return nil, fmt.Errorf("could not list issues: %w", err)
	}

	return issues, nil
}
