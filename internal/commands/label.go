package commands

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
)

func (c *Command) Label() ([]string, error) {
	stat, _ := os.Stdin.Stat()

	ticketID := make([]string, 0)
	labels := make([]string, 0)
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		if len(os.Args) < 3 {
			fmt.Println("Usage: jiwa label <label> <label> ...")
			os.Exit(1)
		}

		in, err := readStdin()
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(bytes.NewBuffer(in))
		for scanner.Scan() {
			ticketID = append(ticketID, StripBaseURL(scanner.Text(), c.Config.BaseURL))
		}
		if scanner.Err() != nil {
			return nil, fmt.Errorf("failed to read in all tickets: %w", err)
		}

		labels = append(labels, os.Args[2:]...)
	} else {
		if len(os.Args) < 4 {
			fmt.Println("Usage: jiwa label <issue ID> <label> <label>...")
			os.Exit(1)
		}
		ticketID = append(ticketID, os.Args[2])
		labels = append(labels, os.Args[3:]...)
	}

	for _, t := range ticketID {
		err := c.Client.LabelIssue(context.TODO(), t, labels...)
		if err != nil {
			return nil, err
		}
	}

	return ticketID, nil
}
