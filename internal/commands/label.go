package commands

import (
	"context"
)

func (c *Command) Label(issues, labels []string) ([]string, error) {
	for _, issue := range issues {
		err := c.Client.LabelIssue(context.TODO(), issue, labels...)
		if err != nil {
			return nil, err
		}
	}

	return issues, nil
}
