package github

import (
	"strings"
	"sync"

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
