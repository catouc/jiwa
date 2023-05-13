package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/andygrunwald/go-jira"
)

type ListInput struct {
	Assignee string
	Project  string
	Status   string
	Labels   []string
}

func (c *Command) List(input ListInput) ([]jira.Issue, error) {
	var user string
	switch input.Assignee {
	case "empty":
		user = "AND assignee is EMPTY"
	case "":
		user = ""
	default:
		user = "AND assignee= \"" + input.Assignee + "\""
	}

	var labelsString string
	if len(input.Labels) != 0 {
		labelsString = "AND labels in (" + strings.Join(input.Labels, ",") + ")"
	}

	project := c.Config.DefaultProject
	if input.Project != "" {
		project = input.Project
	}

	jql := fmt.Sprintf("project=%s AND status=\"%s\" %s %s", project, input.Status, user, labelsString)
	issues, err := c.Client.Search(context.TODO(), jql)
	if err != nil {
		return nil, fmt.Errorf("could not list issues: %w", err)
	}

	return issues, nil
}
