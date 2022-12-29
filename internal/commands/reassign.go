package commands

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
)

func (c *Command) Reassign() ([]string, error) {
	stat, _ := os.Stdin.Stat()

	var user string
	ticketID := make([]string, 0)
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		if len(os.Args) != 3 {
			fmt.Println("Usage: jiwa reassign <username>")
			os.Exit(1)
		}

		in, err := ReadStdin()
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(bytes.NewBuffer(in))
		for scanner.Scan() {
			ticketID = append(ticketID, StripBaseURL(scanner.Text(), c.Config.BaseURL))
		}
		if scanner.Err() != nil {
			return nil, fmt.Errorf("failed to read tickets from stdin: %w", err)
		}

		user = os.Args[2]
	} else {
		if len(os.Args) != 4 {
			fmt.Println("Usage: jiwa reassign <issue ID> <username>")
			os.Exit(1)
		}

		ticketID = append(ticketID, os.Args[2])
		user = os.Args[3]
	}

	for _, t := range ticketID {
		err := c.Client.AssignIssue(context.TODO(), t, user)
		if err != nil {
			return nil, fmt.Errorf("failed to reassign issue %s to %s: %w", t, user, err)
		}
	}

	return ticketID, nil
}
