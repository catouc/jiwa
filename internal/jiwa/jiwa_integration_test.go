// +build integration

package jiwa

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMoveTicketToDOne(t *testing.T) {

	client := Client{
		BaseURL:    "https://catouc.atlassian.net",
		Username:   os.Getenv("JIWA_USERNAME"),
		Password:   os.Getenv("JIWA_PASSWORD"),
		HTTPClient: http.DefaultClient,
		APIVersion: "2",
	}

	// Use jiwa client to create test issue
	issue, err := client.CreateIssue(context.Background(),
		CreateIssueInput{
			Project:     "JIWA",
			Summary:     "TestCase",
			Description: "TestDescription",
			Labels:      []string{"test", "labels"},
			Component:   "TestComponent",
			Assignee:    "atlassian@philipp.boeschen.me",
			Type:        "Task",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func(){
		err := client.DeleteIssue(context.Background(), issue.Key)
		if err != nil {
			t.Fatalf("failed to clean up issue, needs to be manually done: %s", err)
		}
	}()

	err = client.TransitionIssue(context.Background(), issue.Key, "Done")
	if err != nil {
		t.Fatal(err)
	}

	movedIssue, err := client.GetIssue(context.Background(), issue.Key)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, movedIssue.Fields.Status.Name, "Done")
}
