package commands

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
)

func (c *Command) Move() ([]string, error) {
	stat, _ := os.Stdin.Stat()

	ticketID := make([]string, 0)
	var status string
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		if len(os.Args) != 3 {
			fmt.Println("Usage: jiwa mv <status>")
			os.Exit(1)
		}

		in, err := readStdin()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		scanner := bufio.NewScanner(bytes.NewBuffer(in))
		for scanner.Scan() {
			ticketID = append(ticketID, StripBaseURL(scanner.Text(), c.Config.BaseURL))
		}
		if scanner.Err() != nil {
			fmt.Printf("failed to read in all tickets: %s\n", err)
			os.Exit(1)
		}

		status = os.Args[2]
	} else {
		if len(os.Args) != 4 {
			fmt.Println("Usage: jiwa mv <issueID> <status>")
			os.Exit(1)
		}

		ticketID = append(ticketID, os.Args[2])
		status = os.Args[3]
	}

	out := make([]string, 0)
	for _, t := range ticketID {
		err := c.Client.TransitionIssue(context.TODO(), t, status)
		if err != nil {
			return nil, err
		}

		out = append(out, t)
	}
	return ticketID, nil
}
