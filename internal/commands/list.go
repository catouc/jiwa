package commands

import (
	"context"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"strings"
)

func (c *Command) List(userFlag, projectFlag, statusFlag string, labels []string) ([]jira.Issue, error) {
	var user string
	switch userFlag {
	case "empty":
		user = "AND assignee is EMPTY"
	case "":
		user = ""
	default:
		user = "AND assignee= \"" + userFlag + "\""
	}

	var labelsString string
	if len(labels) != 0 {
		labelsString = "AND labels in (" + strings.Join(labels, ",") + ")"
	}

	project := c.Config.DefaultProject
	if projectFlag != "" {
		project = projectFlag
	}

	jql := fmt.Sprintf("project=%s AND status=\"%s\" %s %s", project, statusFlag, user, labelsString)
	issues, err := c.Client.Search(context.TODO(), jql)
	if err != nil {
		return nil, fmt.Errorf("could not list issues: %w", err)
	}

	return issues, nil
}
