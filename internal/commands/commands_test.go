package commands

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommand_ConstructIssueURL(t *testing.T) {
	testData := []struct {
		Name       string
		InCommand  Command
		InIssueKey string
		OutString  string
		OutErr     error
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
			OutErr:     nil,
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
			OutErr:     nil,
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
			OutErr:     errors.New("issueKey must match `^[A-Z]*-[0-9]*$`"),
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
			OutErr:     errors.New("issueKey must match `^[A-Z]*-[0-9]*$`"),
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
			OutErr:     errors.New("issueKey must match `^[A-Z]*-[0-9]*$`"),
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
			OutErr:     errors.New("issueKey must match `^[A-Z]*-[0-9]*$`"),
		},
	}

	for _, td := range testData {
		t.Run(td.Name, func(t *testing.T) {
			t.Parallel()
			result, err := td.InCommand.ConstructIssueURL(td.InIssueKey)

			if td.OutErr == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, td.OutErr.Error(), err.Error())
			}

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
