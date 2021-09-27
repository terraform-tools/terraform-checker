package github

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"

	"github.com/google/go-github/v39/github"
	"github.com/palantir/go-githubapp/githubapp"
)

type CheckHandler struct {
	Client githubapp.ClientCreator
}

func (h *CheckHandler) Handles() []string {
	return []string{"check_run", "check_suite"}
}

func (h *CheckHandler) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	switch eventType {
	case "check_suite":
		return h.startCheckSuite(ctx, payload)
	case "check_run":
		return h.startCheckRun(ctx, payload)
	default:
		return nil
	}
}

func (h *CheckHandler) startCheckSuite(ctx context.Context, payload []byte) error {
	var event github.CheckSuiteEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Print("Error Unmarshal")
		return err
	}

	if event.GetAction() != "requested" {
		return nil
	}

	cs, err := h.checkEventFromSuite(ctx, event)
	if err != nil {
		return err
	}

	cs.allChecks()

	return nil
}

func (h *CheckHandler) startCheckRun(ctx context.Context, payload []byte) error {
	var event github.CheckRunEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Print("Error Unmarshal")
		return err
	}

	if event.GetAction() != "requested_action" {
		return nil
	}

	cs, err := h.checkEventFromRun(ctx, event)
	if err != nil {
		return err
	}

	switch event.GetAction() {
	case "requested_action":
		switch event.GetRequestedAction().Identifier {
		case "fmt":
			return cs.fixFmt()
		default:
			return nil
		}
	default:
		return nil
	}
}
