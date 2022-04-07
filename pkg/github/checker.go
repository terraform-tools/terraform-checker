package github

import (
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
	"github.com/terraform-tools/terraform-checker/pkg/filter"
	"github.com/terraform-tools/terraform-checker/pkg/git"
	"github.com/terraform-tools/terraform-checker/pkg/terraform"
)

func (e *CheckEvent) runChecks(filters ...filter.Option) {
	tfCheckTypes := []string{}
	dirFilter := ""
	for _, f := range filters {
		switch f := f.(type) {
		case *filter.TfCheckTypeFilter:
			if len(f.TfCheckTypes) > 0 {
				tfCheckTypes = f.TfCheckTypes
			}
		case *filter.DirFilter:
			if f.Dir != "" {
				dirFilter = f.Dir
			}
		}
	}

	_, dir, err := git.CloneRepo(e.GetRepo().GetFullName(), e.GetSHA(), e.GetBranch(), e.GetToken())
	if err != nil {
		log.Error().Err(err).Msg("Error cloning the repository")
		return
	}

	var wg sync.WaitGroup

	for _, tfDir := range terraform.FindAllTfDir(dir) {
		tfDir := tfDir

		// If dirFilter is defined and current tfDir does not match, continue
		if dirFilter != "" && !strings.Contains(tfDir, dirFilter) {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			e.processTfDir(dir, tfDir, tfCheckTypes)
		}()
	}
	wg.Wait()
}

func (e *CheckEvent) processTfDir(repoDir, tfDir string, tfCheckTypes []string) {
	relDir := strings.TrimPrefix(strings.ReplaceAll(tfDir, repoDir, ""), "/")
	for _, check := range terraform.GetTfChecks(tfDir, relDir, tfCheckTypes) {
		e.executeCheck(repoDir, check)
	}
}

func (e *CheckEvent) executeCheck(repoDir string, check terraform.TfCheck) {
	cr, _ := e.CreateCheckRun(check)
	checkOk, output := check.Run()
	checkConclusion := check.FailureConclusion()
	if checkOk {
		checkConclusion = githubv4.CheckConclusionStateSuccess
	}
	e.UpdateCheckRun(cr, checkConclusion, output, check)
}
