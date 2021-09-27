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
	_, dir, err := git.CloneRepo(t.Repo.FullName, t.Sha, t.HeadBranch, t.Token)
	if err != nil {
		log.Error().Err(err).Msg("Error cloning the repository")
		return
	}

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
