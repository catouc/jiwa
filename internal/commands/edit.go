package commands

import (
	"context"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"os"
)

func (c *Command) Edit() (string, error) {
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

	summary, description, err := GetIssueIntoEditor(c.Client, ticketID)
	if err != nil {
		return "", fmt.Errorf("failed to get summary and description: %w", err)
	}

	err = c.Client.UpdateIssue(context.TODO(), jira.Issue{
		Key: ticketID,
		Fields: &jira.IssueFields{
			Summary:     summary,
			Description: description,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to update issue: %w", err)
	}

	return ticketID, nil
}
