package commands

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/catouc/jiwa/internal/jiwa"
	"os"
)

func (c *Command) Create(project string, srcFilePath string) (string, error) {
	stat, _ := os.Stdin.Stat()

	var summary, description string
	switch {
	case srcFilePath != "":
		fBytes, err := os.ReadFile(srcFilePath)
		if err != nil {
			fmt.Printf("failed to read file contents: %s", err)
			os.Exit(1)
		}

		scanner := bufio.NewScanner(bytes.NewBuffer(fBytes))

		summary, description, err = BuildSummaryAndDescriptionFromScanner(scanner)
		if err != nil {
			return "", fmt.Errorf("failed to get summary and description: %w", err)
		}
	case (stat.Mode() & os.ModeCharDevice) != 0:
		var err error
		summary, description, err = CreateIssueSummaryDescription("")
		if err != nil {
			return "", fmt.Errorf("failed to get summary and description: %w", err)
		}
	case (stat.Mode() & os.ModeCharDevice) == 0:
		in, err := readStdin()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		scanner := bufio.NewScanner(bytes.NewBuffer(in))
		summary, description, err = BuildSummaryAndDescriptionFromScanner(scanner)
		if err != nil {
			return "", fmt.Errorf("failed to get summary and description: %w", err)
		}
	}

	issue, err := c.Client.CreateIssue(context.TODO(), jiwa.CreateIssueInput{
		Project:     project,
		Summary:     summary,
		Description: description,
		Labels:      nil,
		Type:        "Task",
	})
	if err != nil {
		return "", fmt.Errorf("failed to create issue: %w", err)
	}

	return issue.Key, nil
}
