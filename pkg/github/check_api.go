package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v43/github"
	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
	"github.com/terraform-tools/terraform-checker/pkg/terraform"
)

const checkRunNamePrefix = "terraform-check "

func (e *CheckEvent) CreateAggregatedCheckRun(checkType terraform.TfCheckType) (GhCheckRun, error) {
	checkRunName := fmt.Sprintf("%s%v", checkRunNamePrefix, checkType.String())
	log.Info().Msgf("Create check run %s on repo %s PR %s", checkRunName, e.GetRepo().GetFullName(), e.GetPRURL())

	cr, _, err := e.GetGhClient().Checks.CreateCheckRun(context.TODO(),
		e.GetRepo().GetOwner().GetLogin(),
		e.GetRepo().GetName(),
		github.CreateCheckRunOptions{
			Name:      checkRunName,
			HeadSHA:   e.GetSHA(),
			Status:    github.String(strings.ToLower(string(githubv4.CheckStatusStateInProgress))),
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

func (e *CheckEvent) UpdateAggregatedCheckRun(cr GhCheckRun, checks []terraform.TfCheck) {
	checkRunState := githubv4.CheckConclusionStateSuccess
	annotations := []*github.CheckRunAnnotation{}
	outputText := ""
	actions := []*github.CheckRunAction{}

	for _, check := range checks {
		if !check.IsOK() {
			checkRunState = githubv4.CheckConclusionStateFailure
			actions = append(actions, check.FixActions()...)
		}

		annotations = append(annotations, check.Annotations()...)
		if currentOutput := check.Output(); currentOutput != "" {
			outputText += fmt.Sprintf("**%s:**\n```shell\n%s\n```\n", check.RelDir(), currentOutput)
		}
	}
	checkStatus := fmt.Sprintf("**Check Status:**  %s", CheckConclusionStateEmoji(checkRunState))

	cro := github.CheckRunOutput{
		Title:       &cr.Name,
		Summary:     &checkStatus,
		Annotations: annotations,
	}
	if outputText != "" {
		cro.Text = &outputText
	}

	log.Info().Msgf("Update check run %s on repo %s PR %s", cr.Name, e.GetRepo().GetFullName(), e.GetPRURL())
	_, _, err := e.GetGhClient().Checks.UpdateCheckRun(context.TODO(),
		e.GetRepo().GetOwner().GetLogin(),
		e.GetRepo().GetName(),
		cr.ID,
		github.UpdateCheckRunOptions{
			Name:        cr.Name,
			Status:      github.String(strings.ToLower(string(githubv4.CheckStatusStateCompleted))),
			Output:      &cro,
			Conclusion:  github.String(strings.ToLower(string(checkRunState))),
			CompletedAt: &github.Timestamp{Time: time.Now()},
			Actions:     actions,
		})
	if err != nil {
		log.Error().Err(err).Msg("Error updating check run")
	}
}
