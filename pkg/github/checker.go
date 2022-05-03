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
	defer git.RemoveRepo(dir)

	// SYNCHRONIZATION
	// Wait group for waiting all tasks to be done in the end
	var tasksDone sync.WaitGroup
	// Chan allowing to run only n goroutines at the same time
	currently_running := make(chan int, e.subFolderParallelism)

	for _, tfDir := range terraform.FindAllTfDir(dir) {
		tfDir := tfDir

		// If dirFilter is defined and current tfDir does not match, continue
		if dirFilter != "" && !strings.Contains(tfDir, dirFilter) {
			continue
		}

		currently_running <- 1 // queue current task
		tasksDone.Add(1)
		go func() {
			defer tasksDone.Done()
			e.processTfDir(dir, tfDir, tfCheckTypes)
			<-currently_running // free up space for next one
		}()
	}
	tasksDone.Wait()
	close(currently_running)
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
