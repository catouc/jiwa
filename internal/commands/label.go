package commands

import (
	"context"
	"fmt"
	"os"
)

func (c *Command) Label() ([]string, error) {
	stat, _ := os.Stdin.Stat()

	labels := make([]string, 0)

	var issues []string
	var err error
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		if len(os.Args) < 3 {
			fmt.Println("Usage: jiwa label <label> <label> ...")
			os.Exit(1)
		}

		issues, err = c.readIssueListFromStdin()
		if err != nil {
			return nil, err
		}

		labels = append(labels, os.Args[2:]...)
	} else {
		if len(os.Args) < 4 {
			fmt.Println("Usage: jiwa label <issue ID> <label> <label>...")
			os.Exit(1)
		}
		issues = []string{os.Args[2]}
		labels = append(labels, os.Args[3:]...)
	}

	for _, i := range issues {
		err = c.Client.LabelIssue(context.TODO(), i, labels...)
		if err != nil {
			return nil, err
		}
	}

	return issues, nil
}
