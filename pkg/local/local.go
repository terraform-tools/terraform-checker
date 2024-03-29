package local

import (
	"fmt"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/terraform-tools/terraform-checker/pkg/filter"
	"github.com/terraform-tools/terraform-checker/pkg/terraform"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func StartLocal(dir string, parallelism uint, checkTypes filter.TfCheckTypeFilter) {
	// SYNCHRONIZATION
	// Wait group for waiting all tasks to be done in the end
	var tasksDone sync.WaitGroup
	// Chan allowing to run only n goroutines at the same time
	currentlyRunning := make(chan int, parallelism)

	tfRepos := terraform.FindAllTfDir(dir)
	if len(tfRepos) == 0 {
		log.Error().Msg(fmt.Sprintf("Could not find any terraform folder under %s", dir))
		return
	}

	checks := make(map[string][]terraform.TfCheck)

	for _, tfDir := range tfRepos {
		tfDir := tfDir

		// If tfDir is not enabled, continue
		if !tfDir.IsEnabled() {
			log.Debug().Msgf("TfDir %s skipped, disabled via configuration", tfDir.Path())
			continue
		}

		currentlyRunning <- 1 // queue current task
		tasksDone.Add(1)
		go func() {
			defer tasksDone.Done()

			relDir := strings.TrimPrefix(strings.ReplaceAll(tfDir.Path(), dir, ""), "/")
			for _, check := range terraform.GetTfChecks(tfDir.Path(), relDir, checkTypes.TfCheckTypes) {
				check.Run()
				if _, ok := checks[check.Dir()]; !ok {
					checks[check.Dir()] = []terraform.TfCheck{}
				}
				checks[check.Dir()] = append(checks[check.Dir()], check)
			}
			<-currentlyRunning // free up space for next one
		}()
	}
	tasksDone.Wait()
	close(currentlyRunning)

	renderOutput(checks)
}

func renderOutput(checks map[string][]terraform.TfCheck) {
	okSuffix := " ✅"
	notOkSuffix := " ❌"

	for dir, checks := range checks {
		dirLine := fmt.Sprintf("--- %s", dir)
		color.Blue(strings.Repeat("-", len(dirLine)))
		color.Blue(dirLine)
		for _, check := range checks {
			checkName := cases.Title(language.Und, cases.NoLower).String(check.Name())
			checkTitlePrefix := "\n-- "
			checkTitle := strings.Repeat("-", len(checkTitlePrefix)+len(checkName)+len(okSuffix))
			checkTitle += checkTitlePrefix

			if check.IsOK() {
				checkTitle += color.GreenString(checkName) + okSuffix
			} else {
				checkTitle += color.RedString(checkName) + notOkSuffix
			}
			fmt.Printf("\n%s\n", checkTitle)

			if out := check.Output(); out != "" {
				fmt.Printf("\n%s", out)
			}
		}
	}
}
