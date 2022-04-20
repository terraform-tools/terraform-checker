package github

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/terraform-tools/terraform-checker/pkg/config"
	"github.com/terraform-tools/terraform-checker/pkg/filter"
	"github.com/terraform-tools/terraform-checker/pkg/terraform"

	"github.com/google/go-github/v43/github"
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

func (h *CheckHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error { //nolint:cyclop
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
		dirFilters, checkTypeFilter = h.computeFilters(ctx, event)

		// If the current event is a requested action, execute it
		if e.GetRequestedAction() != nil {
			switch e.GetRequestedAction().Identifier {
			case "fmt":
				return event.fixFmt(dirFilters)
			default:
				return nil
			}
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
	return err == nil && newEvent.IsValid(), newEvent
}

func (h *CheckHandler) getCheckRunEvent(payload []byte) (bool, *CheckEvent) {
	var event github.CheckRunEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Error().Err(err).Msg("Error unmarshal github payload")
		return false, nil
	}

	newEvent, err := NewCheckEvent(h.Client, CheckRunEvent{&event}, h.Config)
	return err == nil && newEvent.IsValid(), newEvent
}

func (h *CheckHandler) getPullRequestEvent(payload []byte) (bool, *CheckEvent) {
	var event github.PullRequestEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Error().Err(err).Msg("Error unmarshal github payload")
		return false, nil
	}

	newEvent, err := NewCheckEvent(h.Client, PullRequestEvent{&event}, h.Config)
	return err == nil && newEvent.IsValid(), newEvent
}

func (h *CheckHandler) computeFilters(ctx context.Context, event *CheckEvent) (dirFilters []filter.Option, checkTypeFilter []filter.Option) {
	externalID := event.ExternalID()
	dir, checkType, err := decodeExternalID(externalID)
	if err != nil {
		return []filter.Option{}, []filter.Option{}
	}

	return []filter.Option{
			&filter.DirFilter{
				Dir: dir,
			},
		},
		[]filter.Option{
			&filter.TfCheckTypeFilter{
				TfCheckTypes: []string{checkType},
			},
		}
}
