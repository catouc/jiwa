package commands

import (
	"context"
	"fmt"
	"os"
)

func (c *Command) Move() ([]string, error) {
	stat, _ := os.Stdin.Stat()

	var status string
	var issues []string
	var err error
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		if len(os.Args) != 3 {
			fmt.Println("Usage: jiwa mv <status>")
			os.Exit(1)
		}

		issues, err = c.readIssueListFromStdin()
		if err != nil {
			return nil, err
		}

		status = os.Args[2]
	} else {
		if len(os.Args) != 4 {
			fmt.Println("Usage: jiwa mv <issueID> <status>")
			os.Exit(1)
		}

		issues = []string{os.Args[2]}
		status = os.Args[3]
	}

	for _, i := range issues {
		err = c.Client.TransitionIssue(context.TODO(), i, status)
		if err != nil {
			return nil, err
		}
	}

	return issues, nil
}
