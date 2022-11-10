package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/catouc/jiwa/internal/editor"
	"github.com/catouc/jiwa/internal/jiwa"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	create   = flag.NewFlagSet("create", flag.ContinueOnError)
	edit     = flag.NewFlagSet("edit", flag.ContinueOnError)
	list     = flag.NewFlagSet("list", flag.ContinueOnError)
	move     = flag.NewFlagSet("move", flag.ContinueOnError)
	reassign = flag.NewFlagSet("reassign", flag.ContinueOnError)
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: jiwa {create|edit|list|move|reassign}\n")
		os.Exit(1)
	}

	httpClient := http.DefaultClient
	httpClient.Timeout = 3 * time.Second

	c := jiwa.Client{
		Username:   os.Getenv("JIRA_USERNAME"),
		Password:   os.Getenv("JIRA_PASSWORD"),
		BaseURL:    "https://catouc.atlassian.net",
		APIVersion: "2",
		HTTPClient: httpClient,
	}

	switch os.Args[1] {
	case "create":
		summary, description, err := CreateIssueSummaryDescription("")
		if err != nil {
			fmt.Printf("failed to get summary and description: %s\n", err)
			os.Exit(1)
		}

		issue, err := c.CreateIssue(context.TODO(), jiwa.CreateIssueInput{
			Project:     "JIWA",
			Summary:     summary,
			Description: description,
			Labels:      nil,
			Type:        "Task",
		})
		if err != nil {
			fmt.Printf("failed to create issue: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("%s/browse/%s\n", c.BaseURL, issue.Key)
	case "edit":
		if len(os.Args) != 3 {
			fmt.Println("Usage: jiwa edit <issue ID>")
			os.Exit(1)
		}

		summary, description, err := GetIssueIntoEditor(c, os.Args[2])
		if err != nil {
			fmt.Printf("failed to get summary and description: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("summary: %s | description: %s\n", summary, description)

		err = c.UpdateIssue(context.TODO(), jira.Issue{
			Key: os.Args[2],
			Fields: &jira.IssueFields{
				Summary:     summary,
				Description: description,
			},
		})
		if err != nil {
			fmt.Printf("failed to update issue: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("%s/browse/%s\n", c.BaseURL, os.Args[2])
	case "list":
	case "ls":
	case "move":
	case "mv":
	case "reassign":
		if len(os.Args) != 4 {
			fmt.Println("Usage: jiwa reassign <issue ID> <username>")
			os.Exit(1)
		}

		err := c.AssignIssue(context.TODO(), os.Args[2], os.Args[3])
		if err != nil {
			fmt.Printf("failed to assign issue to %s: %s\n", os.Args[2], err)
		}
	}
}

func CreateIssueSummaryDescription(prefill string) (string, string, error) {
	fmt.Printf("prefill: %s\n", prefill)

	scanner, cleanup, err := editor.SetupTmpFileWithEditor(prefill)
	if err != nil {
		return "", "", fmt.Errorf("failed to set up scanner on tmpFile: %w", err)
	}
	defer cleanup()

	var title string
	descriptionBuilder := strings.Builder{}
	for scanner.Scan() {
		fmt.Printf("scanner text: %s\n", scanner.Text())
		if title == "" {
			title = scanner.Text()
			continue
		}
		descriptionBuilder.WriteString(scanner.Text())
		descriptionBuilder.WriteString("\n")
	}

	err = scanner.Err()
	if err != nil {
		return "", "", fmt.Errorf("scanner failure: %w", err)
	}

	if title == "" {
		return "", "", errors.New("the summary line needs to be filled at least")
	}

	return title, descriptionBuilder.String(), nil
}

func GetIssueIntoEditor(c jiwa.Client, key string) (string, string, error) {
	issue, err := c.GetIssue(context.TODO(), key)
	if err != nil {
		return "", "", err
	}

	return CreateIssueSummaryDescription(issue.Fields.Summary + "\n" + issue.Fields.Description)
}
