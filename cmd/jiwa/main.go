package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/catouc/jiwa/internal/editor"
	"github.com/catouc/jiwa/internal/jiwa"
	"net/http"
	"os"
	"path"
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

type Config struct {
	BaseURL        string `json:"baseURL"`
	APIVersion     string `json:"apiVersion"`
	EndpointPrefix string `json:"endpointPrefix"`
	Username       string `json:"username"`
	Password       string `json:"password"`
}

var cfg Config

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("cannot locate user home dir, is `$HOME` set? Detailed error: %s\n", err)
		os.Exit(1)
	}

	cfgFileLoc := path.Join(homeDir, ".config", "jiwa", "config.json")

	cfgBytes, err := os.ReadFile(cfgFileLoc)
	if err != nil {
		fmt.Printf("cannot locate configuration file, was it created under %s? Detailed error: %s\n", cfgFileLoc, err)
		os.Exit(1)
	}

	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		fmt.Printf("failed to read configuration file: %s\n", err)
		os.Exit(1)
	}

	username, set := os.LookupEnv("JIWA_USERNAME")
	if set {
		cfg.Username = username
	}
	password, set := os.LookupEnv("JIWA_PASSWORD")
	if set {
		cfg.Password = password
	}

	if cfg.Password == "" || cfg.Username == "" || cfg.BaseURL == "" {
		fmt.Printf(`Config is missing important values, \"baseURL\", \"username\" and \"password\" need to be set.
"username" and "password" can be configured through their respective environment variables "JIWA_USERNAME" and "JIWA_PASSWORD".
The configuration file is located at %s
`, cfgFileLoc)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Printf("Usage: jiwa {create|edit|list|move|reassign}\n")
		os.Exit(1)
	}

}

func main() {
	httpClient := http.DefaultClient
	httpClient.Timeout = 3 * time.Second

	c := jiwa.Client{
		Username:   cfg.Username,
		Password:   cfg.Password,
		BaseURL:    cfg.BaseURL + cfg.EndpointPrefix,
		APIVersion: cfg.APIVersion,
		HTTPClient: httpClient,
	}

	switch os.Args[1] {
	case "create":
		if len(os.Args) != 3 {
			fmt.Println("Usage: jiwa create <project key>")
			os.Exit(1)
		}

		summary, description, err := CreateIssueSummaryDescription("")
		if err != nil {
			fmt.Printf("failed to get summary and description: %s\n", err)
			os.Exit(1)
		}

		issue, err := c.CreateIssue(context.TODO(), jiwa.CreateIssueInput{
			Project:     os.Args[2],
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
	scanner, cleanup, err := editor.SetupTmpFileWithEditor(prefill)
	if err != nil {
		return "", "", fmt.Errorf("failed to set up scanner on tmpFile: %w", err)
	}
	defer cleanup()

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
