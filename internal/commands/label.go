package commands

import (
	"context"
	"os"

	flag "github.com/spf13/pflag"
)

func (c *Command) Label(stdinEmpty bool, flagSet *flag.FlagSet) ([]string, error) {
	var labels []string
		labels = os.Args
		labels = os.Args[1:]

	issues, err := c.ParseInput(stdinEmpty, flagSet, 2, `Usage: jiwa label <issue ID> <label> <label> ...
Alternative usage with issue ID from Stdin: jiwa label <label> <label>...
	`)
	if err != nil {
		return nil, err
	}

	for _, issue := range issues {
		err := c.Client.LabelIssue(context.TODO(), issue, labels...)
		if err != nil {
			return nil, err
		}
	}

	return issues, nil
}

