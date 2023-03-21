package commands

import (
	"context"
)

func (c *Command) Comment(issues []string, comment string) ([]string, error) {
	for _, i := range issues {
		err := c.Client.CommentOnIssue(context.TODO(), i, comment)
		if err != nil {
			return nil, err
		}
	}

	return issues, nil
}
