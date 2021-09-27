package github

import (
	"context"
	"time"

	"github.com/google/go-github/v39/github"
	"github.com/rs/zerolog/log"
)

func (t *CheckEvent) UpdateCheckRun(cr GhCheckRun, status bool, message string) {
	var conclusion string
	if status {
		conclusion = "success"
	} else {
		conclusion = "failure"
	}

	cro := github.CheckRunOutput{
		Title:   github.String("Hello world"),
		Summary: &message,
		Text:    github.String(""),
	}

	fmtAction := github.CheckRunAction{
		Label:       "fmt",
		Description: "format the terraform code",
		Identifier:  "fmt",
	}
	actions := []*github.CheckRunAction{&fmtAction}
	_, _, err := t.GhClient.Checks.UpdateCheckRun(context.TODO(),
		t.Repo.Owner,
		t.Repo.Name,
		cr.ID,
		github.UpdateCheckRunOptions{
			Name:        cr.Name,
			Status:      github.String("completed"),
			Output:      &cro,
			Conclusion:  &conclusion,
			CompletedAt: &github.Timestamp{Time: time.Now()},
			Actions:     actions,
		})
	if err != nil {
		log.Fatal().Err(err).Msg("Error updating check run")
	}
}
