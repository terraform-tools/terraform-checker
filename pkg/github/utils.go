package github

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/terraform-tools/terraform-checker/pkg/terraform"
)

const externalIDParts = 2

func encodeExternalID(e terraform.TfCheck) string {
	return fmt.Sprintf("%s:%s", e.RelDir(), e.Name())
}

func decodeExternalID(id string) (dir string, checkType string, err error) {
	split := strings.Split(id, ":")

	if len(split) != externalIDParts {
		log.Error().Msgf("there should be two parts in ExternalID %v", id)
	}

	dir = split[0]
	checkType = split[1]
	return
}

func getAuthorizedCheckSuiteActions() []string {
	return []string{"requested", "rerequested"}
}

func getAuthorizedCheckRunActions() []string {
	return []string{"requested_action", "rerequested"}
}

func getAuthorizedPullRequestActions() []string {
	return []string{"opened", "reopened"}
}
