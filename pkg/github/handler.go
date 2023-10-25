package github

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/terraform-tools/terraform-checker/pkg/config"
	"github.com/terraform-tools/terraform-checker/pkg/filter"
	"github.com/terraform-tools/terraform-checker/pkg/terraform"

	"github.com/google/go-github/v56/github"
	"github.com/palantir/go-githubapp/githubapp"
)

type CheckHandler struct {
	Client githubapp.ClientCreator
	Config *config.Config
}

func (h *CheckHandler) Init() {
	terraform.InitTfLint()
}

func (h *CheckHandler) Handles() []string {
	return []string{"check_run", "check_suite", "pull_request"}
}

func (h *CheckHandler) Handle(_ context.Context, eventType, _ string, payload []byte) error { //nolint:cyclop
	var ok bool
	var event *CheckEvent
	dirFilters := []filter.Option{}
	checkTypeFilter := []filter.Option{}

	switch eventType {
	case "check_suite":
		ok, event = h.getCheckSuiteEvent(payload)
	case "pull_request":
		ok, event = h.getPullRequestEvent(payload)
	case "check_run":
		ok, event = h.getCheckRunEvent(payload)
	default:
		return nil
	}

	// If current event is not valid, return
	if !ok {
		return nil
	}

	switch e := event.GenericGithubEvent.(type) {
	case CheckRunEvent:
		// If the current event is a CheckRun, we need to compute filters
		// dirFilters, checkTypeFilter = h.computeFilters(ctx, event)

		// If the current event is a requested action, execute it
		if e.GetRequestedAction() != nil {
			switch e.GetRequestedAction().Identifier {
			case terraform.Fmt.String():
				return event.fixFmt()
			default:
			}
		}

		if e.GetAction() == "rerequested" {
			checkTypeFilter = append(checkTypeFilter, &filter.TfCheckTypeFilter{TfCheckTypes: []string{strings.ReplaceAll(e.GetCheckRun().GetName(), checkRunNamePrefix, "")}})
		}

	default:
	}

	event.runChecks(append(dirFilters, checkTypeFilter...)...)
	return nil
}

func (h *CheckHandler) getCheckSuiteEvent(payload []byte) (bool, *CheckEvent) {
	var event github.CheckSuiteEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Error().Err(err).Msg("Error unmarshal github payload")
		return false, nil
	}

	newEvent, err := NewCheckEvent(h.Client, CheckSuiteEvent{&event}, h.Config)

	return err == nil && newEvent.IsValid(h.Config), newEvent
}

func (h *CheckHandler) getCheckRunEvent(payload []byte) (bool, *CheckEvent) {
	var event github.CheckRunEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Error().Err(err).Msg("Error unmarshal github payload")
		return false, nil
	}

	newEvent, err := NewCheckEvent(h.Client, CheckRunEvent{&event}, h.Config)
	return err == nil && newEvent.IsValid(h.Config), newEvent
}

func (h *CheckHandler) getPullRequestEvent(payload []byte) (bool, *CheckEvent) {
	var event github.PullRequestEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Error().Err(err).Msg("Error unmarshal github payload")
		return false, nil
	}

	newEvent, err := NewCheckEvent(h.Client, PullRequestEvent{&event}, h.Config)
	return err == nil && newEvent.IsValid(h.Config), newEvent
}
