package commands

import (
	"context"
)

func (c *Command) Move(issues []string, status string) ([]string, error) {
	for _, i := range issues {
		err := c.Client.TransitionIssue(context.TODO(), i, status)
		if err != nil {
			return nil, err
		}
	}

	return issues, nil
}
