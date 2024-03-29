package commands

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/catouc/jiwa/internal/editor"
	"github.com/catouc/jiwa/internal/jiwa"
)

type Command struct {
	Config Config
	Client jiwa.Client
}

type Config struct {
	BaseURL        string        `json:"baseURL"`
	APIVersion     string        `json:"apiVersion"`
	EndpointPrefix string        `json:"endpointPrefix"`
	Username       string        `json:"username"`
	Password       string        `json:"password"`
	Token          string        `json:"token"`
	Timeout        time.Duration `json:"timeout"`
	DefaultProject string        `json:"defaultProject"`
}

func (c *Config) IsValid() bool {
	switch {
	case c.BaseURL == "":
		return false
	case c.Username == "":
		return false
	case c.Token == "" && c.Password == "":
		return false
	default:
		return true
	}
}

func (c *Config) ReturnCleanEndpointPrefix() string {
	if c.EndpointPrefix != "" && strings.HasPrefix(c.EndpointPrefix, "/") {
		c.EndpointPrefix = strings.TrimPrefix(c.EndpointPrefix, "/")
	}

	return c.EndpointPrefix
}

func BuildCommentFromScanner(scanner *bufio.Scanner) (string, error) {
	commentBuilder := strings.Builder{}
	for scanner.Scan() {
		commentBuilder.WriteString(scanner.Text())
		commentBuilder.WriteString("\n")
	}
	return commentBuilder.String(), scanner.Err()
}

// CreateIssueSummaryDescription takes care of creating an empty tmp file
// and opening an editor on that, reading the result once the editor is closed
// then shoving that into a title and a description.
// SetupTmpFileWithEditor is what you're looking for to just get the file
// thing.
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

func ReadStdin() ([]byte, error) {
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

func (c *Command) StripBaseURL(url string) string {
	issueRegEx, _ := regexp.Compile("^[A-Z]*-[0-9]*$")

	if issueRegEx.MatchString(url) {
		return url
	}

	urlSplit := strings.Split(url, "/browse/")
	if len(urlSplit) != 2 {
		return ""
	}

	return urlSplit[1]
}

func (c *Command) FishOutProject(projectFlag string) (string, error) {
	switch {
	case projectFlag != "":
		return projectFlag, nil
	case projectFlag == "" && c.Config.DefaultProject != "":
		return c.Config.DefaultProject, nil
	default:
		return "", errors.New("either \"defaultProject\" needs to be set in the config or \"--project\" needs to be passed")
	}
}

func (c *Command) ReadIssueListFromStdin() ([]string, error) {
	in, err := ReadStdin()
	if err != nil {
		return nil, err
	}

	issues := make([]string, 0)
	scanner := bufio.NewScanner(bytes.NewBuffer(in))
	for scanner.Scan() {
		issues = append(issues, c.StripBaseURL(scanner.Text()))
	}
	if scanner.Err() != nil {
		return nil, fmt.Errorf("failed to read in all tickets: %w", err)
	}

	return issues, nil
}

func (c *Command) ConstructIssueURL(issueKey string) string {
	// this cannot ever fail to compile? Unless I'm mistaken here...
	issueRegEx, _ := regexp.Compile("^[A-Z]*-[0-9]*$")

	if !issueRegEx.MatchString(issueKey) {
		return ""
	}

	path, err := url.JoinPath(c.Config.BaseURL, c.Config.EndpointPrefix, "browse", issueKey)
	if err != nil {
		return ""
	}

	return path
}
