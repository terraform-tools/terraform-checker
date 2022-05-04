package github

import (
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
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
	if len(tfCheckTypes) == 0 {
		tfCheckTypes = terraform.AllTfCheckTypes()
	}

	_, dir, err := git.CloneRepo(e.GetRepo().GetFullName(), e.GetSHA(), e.GetBranch(), e.GetToken())
	if err != nil {
		log.Error().Err(err).Msg("Error cloning the repository")
		return
	}
	defer git.RemoveRepo(dir)

	// Create CheckRuns
	checkRunMap := e.createCheckRuns(tfCheckTypes)

	// Execute checks
	checks := e.executeChecks(dir, dirFilter, tfCheckTypes)

	// Update CheckRuns
	e.updateCheckRuns(checkRunMap, checks)
}

func (e *CheckEvent) createCheckRuns(tfCheckTypes []string) map[string]GhCheckRun {
	checkRunMap := make(map[string]GhCheckRun, len(tfCheckTypes))
	for _, checkType := range tfCheckTypes {
		newCr, err := e.CreateAggregatedCheckRun(terraform.TfCheckTypeFromString(checkType))
		if err != nil {
			log.Error().Err(err).Msg("there was a problem while creating check_run")
			continue
		}
		checkRunMap[checkType] = newCr
	}
	return checkRunMap
}

func (e *CheckEvent) updateCheckRuns(checkRunMap map[string]GhCheckRun, checks []terraform.TfCheck) {
	for checkType, checkRun := range checkRunMap {
		currentChecks := []terraform.TfCheck{}

		for _, check := range checks {
			if check.Type().String() == checkType {
				currentChecks = append(currentChecks, check)
			}
		}

		e.UpdateAggregatedCheckRun(checkRun, currentChecks)
	}
}

func (e *CheckEvent) executeChecks(dir, dirFilter string, tfCheckTypes []string) (checks []terraform.TfCheck) {
	// SYNCHRONIZATION
	// Wait group for waiting all tasks to be done in the end
	var tasksDone sync.WaitGroup
	// Chan allowing to run only n goroutines at the same time
	currentlyRunning := make(chan int, e.subFolderParallelism)

	for _, tfDir := range terraform.FindAllTfDir(dir) {
		tfDir := tfDir

		// If dirFilter is defined and current tfDir does not match, continue
		if dirFilter != "" && !strings.Contains(tfDir, dirFilter) {
			continue
		}

		currentlyRunning <- 1 // queue current task
		tasksDone.Add(1)
		go func() {
			defer tasksDone.Done()

			relDir := strings.TrimPrefix(strings.ReplaceAll(tfDir, dir, ""), "/")
			for _, check := range terraform.GetTfChecks(tfDir, relDir, tfCheckTypes) {
				check.Run()
				checks = append(checks, check)
			}
			<-currentlyRunning // free up space for next one
		}()
	}
	tasksDone.Wait()
	close(currentlyRunning)
	return checks
}
