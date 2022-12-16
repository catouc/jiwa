package commands

import (
	"context"
	"fmt"
	"os"
)

func (c *Command) Cat() (string, error) {
	stat, _ := os.Stdin.Stat()

	var ticketID string
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		in, err := readStdin()
		if err != nil {
			return "", fmt.Errorf("failed to read stdin: %w", err)
		}

		ticketID = StripBaseURL(string(in), c.Config.BaseURL)
	} else {
		if len(os.Args) != 3 {
			fmt.Println("Usage: jiwa edit <issue ID>")
			os.Exit(1)
		}

		ticketID = os.Args[2]
	}

	issue, err := c.Client.GetIssue(context.TODO(), ticketID)
	if err != nil {
		return "", err
	}

	return issue.Fields.Summary + "\n" + issue.Fields.Description, nil
}
