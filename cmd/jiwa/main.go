package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/catouc/jiwa/internal/commands"
	"github.com/catouc/jiwa/internal/editor"
	"github.com/catouc/jiwa/internal/jiwa"
	flag "github.com/spf13/pflag"
	"net/http"
	"os"
	"path"
	"strings"
	"text/tabwriter"
	"time"
)

var (
	create   = flag.NewFlagSet("create", flag.ContinueOnError)
	edit     = flag.NewFlagSet("edit", flag.ContinueOnError)
	list     = flag.NewFlagSet("list", flag.ContinueOnError)
	move     = flag.NewFlagSet("move", flag.ContinueOnError)
	reassign = flag.NewFlagSet("reassign", flag.ContinueOnError)
	label    = flag.NewFlagSet("label", flag.ContinueOnError)

	createProject = create.StringP("project", "p", "", "Set the project to create the ticket in, if not set it will default to your configured \"defaultProject\"")
	createIn      = create.StringP("in", "i", "", "Control from where the ticket is filled in, can be a file path or \"-\" for stdin")

	listUser    = list.StringP("user", "u", "", "Set the user name to use in the list call, use \"empty\" to list unassigned tickets")
	listStatus  = list.StringP("status", "s", "to do", "Set the status of the tickets you want to see")
	listProject = list.StringP("project", "p", "", "Set the project to search in")
	listOut     = list.StringP("output", "o", "raw", "Set the output to be either \"raw\" for piping or \"table\" for nice formatting")
)

var cfg commands.Config

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
	token, set := os.LookupEnv("JIWA_TOKEN")
	if set {
		cfg.Token = token
	}

	valid := cfg.IsValid()
	if !valid {
		fmt.Printf(`Config is missing important values, \"baseURL\" and \"username\" + \"password\" or \"token\" need to be set.
"username", "password" and "token" can be configured through their respective environment variables "JIWA_USERNAME", "JIWA_PASSWORD" and "JIWA_TOKEN".
The configuration file is located at %s
`, cfgFileLoc)
		os.Exit(1)
	}

	if cfg.APIVersion == "" {
		cfg.APIVersion = "2"
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}

	if len(os.Args) < 2 {
		fmt.Printf("Usage: jiwa {create|edit|list|move|reassign}\n")
		os.Exit(1)
	}

}

func main() {
	httpClient := http.DefaultClient
	httpClient.Timeout = cfg.Timeout

	c := jiwa.Client{
		Username:   cfg.Username,
		Password:   cfg.Password,
		Token:      cfg.Token,
		BaseURL:    cfg.BaseURL + cfg.EndpointPrefix,
		APIVersion: cfg.APIVersion,
		HTTPClient: httpClient,
	}

	cmd := commands.Command{Client: c, Config: cfg}

	stat, _ := os.Stdin.Stat()

	switch os.Args[1] {
	case "create":
		err := create.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("Usage: jiwa create [-project]")
			os.Exit(1)
		}

		var project string
		switch {
		case *createProject == "" && cfg.DefaultProject != "":
			project = cfg.DefaultProject
		case *createProject == "" && cfg.DefaultProject == "":
			fmt.Println("Usage: jiwa create [-project]")
			os.Exit(1)
		case *createProject != "":
			project = *createProject
		}

		key, err := cmd.Create(project, *createIn)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(ConstructIssueURL(key, cfg.BaseURL))
	case "edit":
		key, err := cmd.Edit()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(ConstructIssueURL(key, cfg.BaseURL))
	case "list":
	case "ls":
		err := list.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("Usage: jiwa ls [-user|-status]")
			os.Exit(1)
		}

		var user string
		switch *listUser {
		case "empty":
			user = "AND assignee is EMPTY"
		case "":
			user = ""
		default:
			user = "AND assignee= \"" + *listUser + "\""
		}

		project := cfg.DefaultProject
		if *listProject != "" {
			project = *listProject
		}

		jql := fmt.Sprintf("project=%s AND status=\"%s\" %s", project, *listStatus, user)
		issues, err := c.Search(context.TODO(), jql)
		if err != nil {
			fmt.Printf("could not list issues: %s\n", err)
			os.Exit(1)
		}

		switch *listOut {
		case "raw":
			for _, i := range issues {
				fmt.Println(ConstructIssueURL(i.Key, cfg.BaseURL))
			}
		case "table":
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
			fmt.Fprintf(w, "ID\tSummary\tURL\n")
			for _, i := range issues {
				issueURL := fmt.Sprintf("%s/browse/%s", c.BaseURL, i.Key)
				fmt.Fprintf(w, "%s\t%s\t%s\n", i.Key, i.Fields.Summary, issueURL)
			}
			w.Flush()
		default:
			fmt.Printf("Usage: jiwa ls --out [table|raw]")
		}
	case "move":
	case "mv":
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
				ticketID = append(ticketID, StripBaseURL(scanner.Text(), cfg.BaseURL))
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

		for _, t := range ticketID {
			err := c.TransitionIssue(context.TODO(), t, status)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Println(ConstructIssueURL(t, cfg.BaseURL))
		}
	case "reassign":
		var user string
		ticketID := make([]string, 0)
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			if len(os.Args) != 3 {
				fmt.Println("Usage: jiwa reassign <username>")
				os.Exit(1)
			}

			in, err := readStdin()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			scanner := bufio.NewScanner(bytes.NewBuffer(in))
			for scanner.Scan() {
				ticketID = append(ticketID, StripBaseURL(scanner.Text(), cfg.BaseURL))
			}
			if scanner.Err() != nil {
				fmt.Printf("failed to read in all tickets: %s\n", err)
				os.Exit(1)
			}

			ticketID = append(ticketID, StripBaseURL(string(in), cfg.BaseURL))
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
			err := c.AssignIssue(context.TODO(), t, user)
			if err != nil {
				fmt.Printf("failed to assign issue to %s: %s\n", t, err)
				os.Exit(1)
			}

			fmt.Println(ConstructIssueURL(t, cfg.BaseURL))
		}
	case "label":
		ticketID := make([]string, 0)
		labels := make([]string, 0)
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			if len(os.Args) < 3 {
				fmt.Println("Usage: jiwa label <label> <label> ...")
			}

			in, err := readStdin()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			scanner := bufio.NewScanner(bytes.NewBuffer(in))
			for scanner.Scan() {
				ticketID = append(ticketID, StripBaseURL(scanner.Text(), cfg.BaseURL))
			}
			if scanner.Err() != nil {
				fmt.Printf("failed to read in all tickets: %s\n", err)
				os.Exit(1)
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
			err := c.LabelIssue(context.TODO(), t, labels...)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Println(ConstructIssueURL(t, cfg.BaseURL))
		}
	}
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

func StripBaseURL(url, baseURL string) string {
	return strings.TrimPrefix(strings.TrimSpace(url), baseURL+"/browse/")
}

func ConstructIssueURL(issueKey, baseURL string) string {
	return fmt.Sprintf("%s/browse/%s", baseURL, issueKey)
}
