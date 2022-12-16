package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/catouc/jiwa/internal/commands"
	"github.com/catouc/jiwa/internal/jiwa"
	flag "github.com/spf13/pflag"
	"net/http"
	"os"
	"path"
	"text/tabwriter"
	"time"
)

var (
	create    = flag.NewFlagSet("create", flag.ContinueOnError)
	edit      = flag.NewFlagSet("edit", flag.ContinueOnError)
	list      = flag.NewFlagSet("list", flag.ContinueOnError)
	move      = flag.NewFlagSet("move", flag.ContinueOnError)
	reassign  = flag.NewFlagSet("reassign", flag.ContinueOnError)
	label     = flag.NewFlagSet("label", flag.ContinueOnError)
	issueType = flag.NewFlagSet("issue-type", flag.ContinueOnError)

	createProject    = create.StringP("project", "p", "", "Set the project to create the ticket in, if not set it will default to your configured \"defaultProject\"")
	createFile       = create.StringP("file", "f", "", "Point to a file that contains your ticket")
	createTicketType = create.StringP("ticket-type", "t", "Task", "Sets the type of ticket to open, defaults to \"Task\"")

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
		fmt.Printf("Usage: jiwa {create|edit|ls|mv|reassign}\n")
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
		BaseURL:    cfg.BaseURL + "/" + cfg.ReturnCleanEndpointPrefix(),
		APIVersion: cfg.APIVersion,
		HTTPClient: httpClient,
	}

	cmd := commands.Command{Client: c, Config: cfg}

	switch os.Args[1] {
	case "create":
		err := create.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("Usage: jiwa create [-project]")
			os.Exit(1)
		}

		project, err := cmd.FishOutProject(*createProject)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		key, err := cmd.Create(project, *createFile, *createTicketType)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(ConstructIssueURL(key, cfg.BaseURL, cfg.ReturnCleanEndpointPrefix()))
	case "edit":
		key, err := cmd.Edit()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(ConstructIssueURL(key, cfg.BaseURL, cfg.ReturnCleanEndpointPrefix()))
	case "list":
		err := list.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("Usage: jiwa ls [-user|-status]")
			os.Exit(1)
		}

		issues, err := cmd.List(*listUser, *listProject, *listStatus)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		switch *listOut {
		case "raw":
			for _, i := range issues {
				fmt.Println(ConstructIssueURL(i.Key, cmd.Config.BaseURL, cfg.ReturnCleanEndpointPrefix()))
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
	case "ls":
		err := list.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("Usage: jiwa ls [-user|-status]")
			os.Exit(1)
		}

		issues, err := cmd.List(*listUser, *listProject, *listStatus)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		switch *listOut {
		case "raw":
			for _, i := range issues {
				fmt.Println(ConstructIssueURL(i.Key, cmd.Config.BaseURL, cfg.ReturnCleanEndpointPrefix()))
			}
		case "table":
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
			fmt.Fprintf(w, "ID\tSummary\tURL\n")
			for _, i := range issues {
				issueURL := ConstructIssueURL(i.Key, cmd.Config.BaseURL, cfg.ReturnCleanEndpointPrefix())
				fmt.Fprintf(w, "%s\t%s\t%s\n", i.Key, i.Fields.Summary, issueURL)
			}
			w.Flush()
		default:
			fmt.Printf("Usage: jiwa ls --out [table|raw]")
		}

	case "move":
		issues, err := cmd.Move()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, i := range issues {
			fmt.Println(ConstructIssueURL(i, cmd.Config.BaseURL, cfg.ReturnCleanEndpointPrefix()))
		}
	case "mv":
		issues, err := cmd.Move()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, i := range issues {
			fmt.Println(ConstructIssueURL(i, cmd.Config.BaseURL, cfg.ReturnCleanEndpointPrefix()))
		}
	case "reassign":
		issues, err := cmd.Reassign()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, i := range issues {
			fmt.Println(ConstructIssueURL(i, cmd.Config.BaseURL, cfg.ReturnCleanEndpointPrefix()))
		}
	case "label":
		issues, err := cmd.Label()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, i := range issues {
			fmt.Println(ConstructIssueURL(i, cmd.Config.BaseURL, cfg.ReturnCleanEndpointPrefix()))
		}
	case "issue-type":
		if len(os.Args) < 3 {
			fmt.Println("Usage: jiwa issue-type <project-key>")
			os.Exit(1)
		}

		project, err := c.GetProject(context.TODO(), os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, it := range project.IssueTypes {
			fmt.Println(it.Name)
		}
	}
}

func ConstructIssueURL(issueKey, baseURL, endpointPrefix string) string {
	switch endpointPrefix {
	case "":
		return fmt.Sprintf("%s/browse/%s", baseURL, issueKey)
	default:
		return fmt.Sprintf("%s/%s/browse/%s", endpointPrefix, baseURL, issueKey)
	}

}
