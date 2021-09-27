package github

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v39/github"
	"github.com/rs/zerolog/log"
	"github.com/terraform-tools/terraform-checker/pkg/git"
	"github.com/terraform-tools/terraform-checker/pkg/terraform"
)

func (t *CheckEvent) allChecks() {
	_, dir, _ := git.CloneRepo(t.Repo.FullName, t.Sha, t.HeadBranch, t.Token)

	var wg sync.WaitGroup

	for _, tfDir := range terraform.FindAllTfDir(dir) {
		tfDir := tfDir
		wg.Add(1)
		go func() {
			defer wg.Done()
			t.checkOneTfDir(dir, tfDir)
		}()
	}
	wg.Wait()
}

func (t *CheckEvent) checkOneTfDir(repoDir, tfDir string) {
	relDir := strings.Replace(tfDir, repoDir, "", 1)
	cr, _ := t.CreateCheckRun(relDir)
	log.Info().Msgf("Checking tfdir %v", tfDir)
	checkOk, msg := terraform.CheckTfDir(tfDir)
	log.Print("check output ", checkOk)
	t.UpdateCheckRun(cr, checkOk, msg)
}

func (t *CheckEvent) CreateCheckRun(dir string) (GhCheckRun, error) {
	log.Print("Create check run ", dir)
	cr, _, err := t.GhClient.Checks.CreateCheckRun(context.TODO(),
		t.Repo.Owner,
		t.Repo.Name,
		github.CreateCheckRunOptions{
			Name:      fmt.Sprintf("tf %v", dir),
			HeadSHA:   t.Sha,
			Status:    github.String("in_progress"),
			StartedAt: &github.Timestamp{Time: time.Now()},
		})
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating check run")

		return GhCheckRun{}, err
	}
	return GhCheckRun{
		Name: fmt.Sprintf("tf %v", dir),
		ID:   *cr.ID,
	}, nil
}

func (t *CheckEvent) UpdateCheckRun(cr GhCheckRun, status bool, message string) {
	var conclusion string
	if status {
		conclusion = "success"
	} else {
		conclusion = "failure"
	}

	cro := github.CheckRunOutput{
		Title:   github.String("output"),
		Summary: github.String("output summary"),
		Text:    &message,
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
