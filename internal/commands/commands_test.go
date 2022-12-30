package commands

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestCommand_ConstructIssueURL(t *testing.T) {
	testData := []struct {
		Name       string
		InCommand  Command
		InIssueKey string
		OutString  string
	}{
		{
			Name: "HappyPathWithEndpointPrefix",
			InCommand: Command{
				Config: Config{
					BaseURL:        "https://catouc.atlassian.net",
					APIVersion:     "2",
					EndpointPrefix: "/jira",
				},
			},
			InIssueKey: "JIWA-001",
			OutString:  "https://catouc.atlassian.net/jira/browse/JIWA-001",
		},
		{
			Name: "HappyPath",
			InCommand: Command{
				Config: Config{
					BaseURL:        "https://catouc.atlassian.net",
					APIVersion:     "2",
					EndpointPrefix: "",
				},
			},
			InIssueKey: "JIWA-001",
			OutString:  "https://catouc.atlassian.net/browse/JIWA-001",
		},
		{
			Name: "EmptyIssueKey",
			InCommand: Command{
				Config: Config{
					BaseURL:        "https://catouc.atlassian.net",
					APIVersion:     "2",
					EndpointPrefix: "",
				},
			},
			InIssueKey: "",
			OutString:  "",
		},
		{
			Name: "EmptyIssueKeyWithEndpointPrefix",
			InCommand: Command{
				Config: Config{
					BaseURL:        "https://catouc.atlassian.net",
					APIVersion:     "2",
					EndpointPrefix: "/jira",
				},
			},
			InIssueKey: "",
			OutString:  "",
		},
		{
			Name: "InvalidIssueKey",
			InCommand: Command{
				Config: Config{
					BaseURL:        "https://catouc.atlassian.net",
					APIVersion:     "2",
					EndpointPrefix: "",
				},
			},
			InIssueKey: "01Something",
			OutString:  "",
		},
		{
			Name: "InvalidIssueKeyWithEndpointPrefix",
			InCommand: Command{
				Config: Config{
					BaseURL:        "https://catouc.atlassian.net",
					APIVersion:     "2",
					EndpointPrefix: "/jira",
				},
			},
			InIssueKey: "01Something",
			OutString:  "",
		},
	}

	for _, td := range testData {
		t.Run(td.Name, func(t *testing.T) {
			t.Parallel()
			result := td.InCommand.ConstructIssueURL(td.InIssueKey)

			assert.Equal(t, td.OutString, result)
		})
	}
}

func TestCommand_StripBaseURL(t *testing.T) {
	testData := []struct {
		Name      string
		InCommand Command
		InURL     string
		OutString string
	}{
		{
			Name: "HappyPath",
			InCommand: Command{
				Config: Config{
					BaseURL:        "https://catouc.atlassian.net",
					APIVersion:     "2",
					EndpointPrefix: "",
				},
			},
			InURL:     "https://catouc.atlassian.net/browse/JIWA-001",
			OutString: "JIWA-001",
		},
		{
			Name: "HappyPathWithEndpointPrefix",
			InCommand: Command{
				Config: Config{
					BaseURL:        "https://catouc.atlassian.net",
					APIVersion:     "2",
					EndpointPrefix: "/jira",
				},
			},
			InURL:     "https://catouc.atlassian.net/jira/browse/JIWA-001",
			OutString: "JIWA-001",
		},
		{
			Name: "ConfiguredEndpointPrefixButMissingInURL",
			InCommand: Command{
				Config: Config{
					BaseURL:        "https://catouc.atlassian.net",
					APIVersion:     "2",
					EndpointPrefix: "/jira",
				},
			},
			InURL:     "https://catouc.atlassian.net/browse/JIWA-001",
			OutString: "JIWA-001",
		},
		{
			Name: "NotAJiraTicketURL",
			InCommand: Command{
				Config: Config{
					BaseURL:        "https://catouc.atlassian.net",
					APIVersion:     "2",
					EndpointPrefix: "/jira",
				},
			},
			InURL:     "invalidURL",
			OutString: "",
		},
		{
			Name: "JustIssueKey",
			InCommand: Command{
				Config: Config{
					BaseURL:        "https://catouc.atlassian.net",
					APIVersion:     "2",
					EndpointPrefix: "/jira",
				},
			},
			InURL:     "JIWA-001",
			OutString: "JIWA-001",
		},
	}

	for _, td := range testData {
		t.Run(td.Name, func(t *testing.T) {
			t.Parallel()
			result := td.InCommand.StripBaseURL(td.InURL)

			assert.Equal(t, td.OutString, result)
		})
	}
}

type GetIssuesAndArgsFromFlagSetInput struct {
	FlagSet       *flag.FlagSet
	Command       Command
	MinArgsNormal int
	MinArgsStdin  int
	ArgHelp       string
	StdinHelp     string
}

func TestCommand_GetIssuesAndArgsFromFlagSet(t *testing.T) {
	testData := []struct {
		Name      string
		In        GetIssuesAndArgsFromFlagSetInput
		OutIssues []string
		OutArgs   []string
	}{
		{
			Name: "HappyPathSingleIssueNoArgsNoStdin",
			In: GetIssuesAndArgsFromFlagSetInput{
				FlagSet: flag.NewFlagSet("test", flag.ContinueOnError),
				Command: Command{
					Config: Config{
						BaseURL:        "https://catouc.atlassian.net",
						APIVersion:     "2",
						EndpointPrefix: "",
					},
				},
				MinArgsNormal: 1,
				MinArgsStdin:  0,
				ArgHelp:       "help for args",
				StdinHelp:     "help for stdin",
			},
			OutIssues: []string{"JIWA-001"},
			OutArgs:   []string{"cat"},
		},
	}

	for _, td := range testData {
		t.Run(td.Name, func(t *testing.T) {
			t.Parallel()

			cleanupStdin, err := mockStdin(t, "https://catouc.atlassian.net/browse/JIWA-001")
			if err != nil {
				t.Fatal(err)
			}
			defer cleanupStdin()

			cleanupOSArgs := mockOSArgs(t, os.Args[0], "jiwa", "cat")
			defer cleanupOSArgs()

			issues, args := td.In.Command.GetIssuesAndArgsFromFlagSet(
				td.In.FlagSet,
				td.In.MinArgsNormal,
				td.In.MinArgsStdin,
				td.In.ArgHelp,
				td.In.StdinHelp,
			)

			fmt.Println(args)
			assert.ElementsMatch(t, td.OutIssues, issues)
			assert.ElementsMatch(t, td.OutArgs, args)
		})
	}
}

func mockStdin(t *testing.T, input string) (func(), error) {
	t.Helper()

	oldStdin := os.Stdin

	tmpFile, err := os.CreateTemp(t.TempDir(), strings.ReplaceAll(t.Name(), "/", "_"))
	if err != nil {
		return nil, err
	}

	content := []byte(input)

	_, err = tmpFile.Write(content)
	if err != nil {
		return nil, err
	}

	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	os.Stdin = tmpFile

	return func() {
		os.Stdin = oldStdin
		os.Remove(tmpFile.Name())
	}, nil
}

func mockOSArgs(t *testing.T, newArgs ...string) func() {
	oldOSArgs := os.Args
	os.Args = newArgs
	return func() {
		os.Args = oldOSArgs
	}
}
