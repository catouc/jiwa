package commands

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/catouc/jiwa/internal/editor"
	"github.com/catouc/jiwa/internal/jiwa"
	"os"
	"strings"
)

type Command struct {
	Config Config
	Client jiwa.Client
}

type Config struct {
	BaseURL        string `json:"baseURL"`
	APIVersion     string `json:"apiVersion"`
	EndpointPrefix string `json:"endpointPrefix"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	DefaultProject string `json:"defaultProject"`
}

func CreateIssueSummaryDescription(prefill string) (string, string, error) {
	scanner, cleanup, err := editor.SetupTmpFileWithEditor(prefill)
	if err != nil {
		return "", "", fmt.Errorf("failed to set up scanner on tmpFile: %w", err)
	}
	defer cleanup()

	title, description, err := BuildSummaryAndDescriptionFromScanner(scanner)
	if err != nil {
		return "", "", fmt.Errorf("scanner failure: %w", err)
	}

	if title == "" {
		return "", "", errors.New("the summary line needs to be filled at least")
	}

	return title, description, nil
}

func BuildSummaryAndDescriptionFromScanner(scanner *bufio.Scanner) (string, string, error) {
	var title string
	descriptionBuilder := strings.Builder{}
	for scanner.Scan() {
		if title == "" {
			title = scanner.Text()
			continue
		}
		descriptionBuilder.WriteString(scanner.Text())
		descriptionBuilder.WriteString("\n")
	}

	return title, descriptionBuilder.String(), scanner.Err()
}

func GetIssueIntoEditor(c jiwa.Client, key string) (string, string, error) {
	issue, err := c.GetIssue(context.TODO(), key)
	if err != nil {
		return "", "", err
	}

	return CreateIssueSummaryDescription(issue.Fields.Summary + "\n" + issue.Fields.Description)
}

func readStdin() ([]byte, error) {
	var buf []byte
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		buf = append(buf, scanner.Bytes()...)
		buf = append(buf, 10) // add the newline back into the buffer
	}

	err := scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("failed to read stdin: %v", err)
	}

	return buf, nil
}

func (c *Command) Create(project string, srcFilePath string) string {
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
			fmt.Printf("failed to get summary and description: %s\n", err)
			os.Exit(1)
		}
	case (stat.Mode() & os.ModeCharDevice) != 0:
		var err error
		summary, description, err = CreateIssueSummaryDescription("")
		if err != nil {
			fmt.Printf("failed to get summary and description: %s\n", err)
			os.Exit(1)
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
			fmt.Printf("failed to get summary and description: %s\n", err)
			os.Exit(1)
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
		fmt.Printf("failed to create issue: %s\n", err)
		os.Exit(1)
	}

	return issue.Key
}
