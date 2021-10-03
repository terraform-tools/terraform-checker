package github

import (
	"context"
	"fmt"
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
		Title:   github.String("Terraform check is a " + conclusion),
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
		log.Error().Err(err).Msg("Error updating check run")
	}
}

func (t *CheckEvent) CreateCheckRun(dir string) (GhCheckRun, error) {
	checkRunName := fmt.Sprintf("terraform-check %v", dir)
	log.Print("Create check run ", checkRunName)
	cr, _, err := t.GhClient.Checks.CreateCheckRun(context.TODO(),
		t.Repo.Owner,
		t.Repo.Name,
		github.CreateCheckRunOptions{
			Name:      checkRunName,
			HeadSHA:   t.Sha,
			Status:    github.String("in_progress"),
			StartedAt: &github.Timestamp{Time: time.Now()},
		})
	if err != nil {
		log.Error().Err(err).Msg("Error creating check run")

		return GhCheckRun{}, err
	}
	return GhCheckRun{
		Name: checkRunName,
		ID:   *cr.ID,
	}, nil
}
