package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"text/tabwriter"
	"time"

	"github.com/catouc/jiwa/internal/commands"
	"github.com/catouc/jiwa/internal/jiwa"
	flag "github.com/spf13/pflag"
)

var (
	cat       = flag.NewFlagSet("cat", flag.ContinueOnError)
	comment   = flag.NewFlagSet("comment", flag.ContinueOnError)
	create    = flag.NewFlagSet("create", flag.ContinueOnError)
	edit      = flag.NewFlagSet("edit", flag.ContinueOnError)
	issueType = flag.NewFlagSet("issue-type", flag.ContinueOnError)
	label     = flag.NewFlagSet("label", flag.ContinueOnError)
	list      = flag.NewFlagSet("list", flag.ContinueOnError)
	move      = flag.NewFlagSet("move", flag.ContinueOnError)
	reassign  = flag.NewFlagSet("reassign", flag.ContinueOnError)
	search    = flag.NewFlagSet("search", flag.ContinueOnError)

	catComments = cat.BoolP("comments", "c", false, "Toggle to include comments in the printout or not")

	createProject    = create.StringP("project", "p", "", "Set the project to create the ticket in, if not set it will default to your configured \"defaultProject\"")
	createFile       = create.StringP("file", "f", "", "Point to a file that contains your ticket")
	createTicketType = create.StringP("ticket-type", "t", "Task", "Sets the type of ticket to open, defaults to \"Task\"")
	createComponent  = create.StringP("component", "c", "", "Set the component of your ticket")

	listUser    = list.StringP("user", "u", "", "Set the user name to use in the list call, use \"empty\" to list unassigned tickets")
	listStatus  = list.StringP("status", "s", "to do", "Set the status of the tickets you want to see")
	listProject = list.StringP("project", "p", "", "Set the project to search in")
	listOut     = list.StringP("output", "o", "raw", "Set the output to be either \"raw\" for piping or \"table\" for nice formatting")
	listLabels  = list.StringArrayP("label", "l", nil, "Search for specific labels, all labels are joined by an OR")
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
		fmt.Printf("Usage: jiwa {cat|comment|create|edit|issueType||label|list|move|reassign|search}\n")
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

	stat, _ := os.Stdin.Stat()

	switch os.Args[1] {
	case "cat":
		err := cat.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("jiwa cat <issue-id>")
			fmt.Println("echo \"<issue-id>\" | jiwa cat <issue-id>")
			os.Exit(1)
		}

		var issues []string
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			issues, err = cmd.ReadIssueListFromStdin()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			if len(cat.Args()) == 0 {
				fmt.Println("Usage: jiwa cat <issue-id>")
				os.Exit(1)
			}

			issues = []string{cmd.StripBaseURL(cat.Arg(0))}
		}

		issue, err := cmd.Cat(issues[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(issue.Fields.Summary+"\n"+issue.Fields.Description, nil)

		if *catComments {
			for _, comment := range issue.Fields.Comments.Comments {
				fmt.Printf("%s wrote on %s:\n%s\n", comment.Author.Name, comment.Created, comment.Body)
			}
		}
	case "comment":
		err := comment.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("Usage: jiwa comment <issue-id> <comment>")
			fmt.Println("echo \"<issue-id>\" | jiwa comment <comment>")
			os.Exit(1)
		}

		var issues []string
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			if len(comment.Args()) == 0 {
				fmt.Println("echo \"<issue-id>\" | jiwa comment <comment>")
				os.Exit(1)
			}
			issues, err = cmd.ReadIssueListFromStdin()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			if len(comment.Args()) < 2 {
				fmt.Println("Usage: jiwa comment <issue-id> <comment>")
				os.Exit(1)
			}

			issues = []string{cmd.StripBaseURL(comment.Arg(0))}
		}

		commentedIssues, err := cmd.Comment(issues, comment.Arg(1))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, issue := range commentedIssues {
			fmt.Println(cmd.ConstructIssueURL(issue))
		}
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

		key, err := cmd.Create(project, *createFile, *createTicketType, *createComponent)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(cmd.ConstructIssueURL(key))
	case "edit":
		err := edit.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("jiwa edit <issue-id>")
			fmt.Println("echo \"<issue-id>\" | jiwa edit")
			os.Exit(1)
		}

		var issues []string
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			issues, err = cmd.ReadIssueListFromStdin()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			if len(edit.Args()) == 0 {
				fmt.Println("Usage: jiwa edit <issue ID>")
				os.Exit(1)
			}

			issues = []string{cmd.StripBaseURL(edit.Arg(0))}
		}

		key, err := cmd.Edit(issues[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(cmd.ConstructIssueURL(key))
	case "issue-type":
		err := issueType.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("jiwa issue-type <project-key>")
			os.Exit(1)
		}

		if len(issueType.Args()) == 0 {
			fmt.Println("jiwa issue-type <project-key>")
			os.Exit(1)
		}

		issueTypes, err := cmd.IssueTypes(issueType.Arg(0))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, it := range issueTypes {
			fmt.Println(it.Name)
		}
	case "label":
		err := label.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("jiwa label <issue ID> <label> <label>...")
			fmt.Println("echo \"<issue-id>\" | jiwa label <label> <label> ...")
			os.Exit(1)
		}

		var labels []string
		var issues []string
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			if len(label.Args()) == 0 {
				fmt.Println("Usage: jiwa label <label> <label> ...")
				os.Exit(1)
			}

			issues, err = cmd.ReadIssueListFromStdin()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			labels = label.Args()
		} else {
			if len(label.Args()) < 2 {
				fmt.Println("Usage: jiwa label <issue ID> <label> <label>...")
				os.Exit(1)
			}

			issues = []string{cmd.StripBaseURL(label.Arg(0))}
			labels = label.Args()[1:]
		}

		labelledIssues, err := cmd.Label(issues, labels)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, issue := range labelledIssues {
			fmt.Println(cmd.ConstructIssueURL(issue))
		}
	case "list":
		err := list.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("Usage: jiwa list [--user|--status|--project|--label]")
			os.Exit(1)
		}

		issues, err := cmd.List(*listUser, *listProject, *listStatus, *listLabels)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		switch *listOut {
		case "raw":
			for _, i := range issues {
				fmt.Println(cmd.ConstructIssueURL(i.Key))
			}
		case "table":
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
			fmt.Fprintf(w, "ID\tSummary\tURL\n")
			for _, i := range issues {
				fmt.Fprintf(w, "%s\t%s\t%s\n", i.Key, i.Fields.Summary, cmd.ConstructIssueURL(i.Key))
			}
			w.Flush()
		default:
			fmt.Printf("Usage: jiwa ls --out [table|raw]")
		}
	case "ls":
		err := list.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("Usage: jiwa ls [--user|--status|--project|--label]")
			os.Exit(1)
		}

		issues, err := cmd.List(*listUser, *listProject, *listStatus, *listLabels)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		switch *listOut {
		case "raw":
			for _, i := range issues {
				fmt.Println(cmd.ConstructIssueURL(i.Key))
			}
		case "table":
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
			fmt.Fprintf(w, "ID\tSummary\tURL\n")
			for _, i := range issues {
				fmt.Fprintf(w, "%s\t%s\t%s\n", i.Key, i.Fields.Summary, cmd.ConstructIssueURL(i.Key))
			}
			w.Flush()
		default:
			fmt.Printf("Usage: jiwa ls --out [table|raw]")
		}
	case "move":
		err := move.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("jiwa move <issue-id> <status>")
			fmt.Println("echo \"<issue-id>\" | jiwa move <status>")
			os.Exit(1)
		}

		var status string
		var issues []string
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			if len(move.Args()) == 0 {
				fmt.Println("Usage: jiwa move <status>")
				os.Exit(1)
			}

			issues, err = cmd.ReadIssueListFromStdin()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			status = move.Arg(0)
		} else {
			if len(move.Args()) < 2 {
				fmt.Println("Usage: jiwa move <issueID> <status>")
				os.Exit(1)
			}

			issues = []string{cmd.StripBaseURL(move.Arg(0))}
			status = move.Arg(1)
		}

		movedIssues, err := cmd.Move(issues, status)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, issue := range movedIssues {
			fmt.Println(cmd.ConstructIssueURL(issue))
		}
	case "mv":
		err := move.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("jiwa mv <issue-id> <status>")
			fmt.Println("echo \"<issue-id>\" | jiwa mv <status>")
			os.Exit(1)
		}

		var status string
		var issues []string
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			if len(move.Args()) == 0 {
				fmt.Println("Usage: jiwa mv <status>")
				os.Exit(1)
			}

			issues, err = cmd.ReadIssueListFromStdin()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			status = move.Arg(0)
		} else {
			if len(move.Args()) < 2 {
				fmt.Println("Usage: jiwa mv <issueID> <status>")
				os.Exit(1)
			}

			issues = []string{cmd.StripBaseURL(move.Arg(0))}
			status = move.Arg(1)
		}

		movedIssues, err := cmd.Move(issues, status)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, issue := range movedIssues {
			fmt.Println(cmd.ConstructIssueURL(issue))
		}
	case "reassign":
		err := reassign.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("jiwa reassign <issue-id> <username>")
			fmt.Println("echo \"<issue-id>\" | jiwa reassign <username>")
			os.Exit(1)
		}

		var user string
		var issues []string
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			if len(reassign.Args()) == 0 {
				fmt.Println("Usage: jiwa reassign <username>")
				os.Exit(1)
			}

			issues, err = cmd.ReadIssueListFromStdin()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			user = reassign.Arg(0)
		} else {
			if len(reassign.Args()) < 2 {
				fmt.Println("Usage: jiwa reassign <issue ID> <username>")
				os.Exit(1)
			}

			issues = []string{cmd.StripBaseURL(reassign.Arg(0))}
			user = reassign.Arg(1)
		}

		reassignedIssues, err := cmd.Reassign(issues, user)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, issue := range reassignedIssues {
			fmt.Println(cmd.ConstructIssueURL(issue))
		}
	case "search":
		err := search.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("jiwa search \"<jql query>\"")
			os.Exit(1)
		}

		if len(search.Args()) == 0 {
			fmt.Println("jiwa search \"<jql query>\"")
			os.Exit(1)
		}

		issues, err := cmd.Search(search.Arg(0))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, i := range issues {
			fmt.Println(cmd.ConstructIssueURL(i.Key))
		}
	}
}
