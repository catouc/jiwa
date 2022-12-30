package commands

import (
	"github.com/stretchr/testify/assert"
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
