package commands

import (
	"context"
	"fmt"
)

func (c *Command) Reassign(issues []string, username string) ([]string, error) {

	for _, issue := range issues {
		err := c.Client.AssignIssue(context.TODO(), issue, username)
		if err != nil {
			return nil, fmt.Errorf("failed to reassign issue %s to %s: %w", issue, username, err)
		}
	}

	return issues, nil
}
