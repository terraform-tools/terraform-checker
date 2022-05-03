package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v43/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rs/zerolog/log"
	"github.com/terraform-tools/terraform-checker/pkg/config"
	"github.com/terraform-tools/terraform-checker/pkg/errors"
	"github.com/terraform-tools/terraform-checker/pkg/utils"
)

type Repo struct {
	*github.Repository
}

func (r *Repo) HasTopic(t string) bool {
	return utils.StrInSlice(r.Topics, t)
}

func (r *Repo) IsValid(config *config.Config) (ok bool, err error) {
	if len(config.GHRepoWhitelist) > 0 {
		if ok = utils.StrInSlice(config.GHRepoWhitelist, r.GetName()); !ok {
			err = errors.RepoNotValidError(fmt.Sprintf("skipped repo %s because it's not in whitelist", r.GetFullName()))
			log.Debug().Err(err).Msg("")
			return
		}
	}

	if !r.HasTopic(config.GHRepoTopic) {
		err = errors.RepoNotValidError(fmt.Sprintf("skipped repo %s because it does not have topic %s", r.GetFullName(), config.GHRepoTopic))
		log.Debug().Err(err).Msg("")
		return
	}
	return true, nil
}

type GhCheckRun struct {
	Name string
	ID   int64
}

type CheckEvent struct {
	GenericGithubEvent
	repo                 Repo
	sha                  string
	token                string
	branch               string
	prURL                string
	ghClient             *github.Client
	subFolderParallelism int
}

func (c *CheckEvent) GetRepo() *Repo {
	return &c.repo
}

func (c *CheckEvent) GetSHA() string {
	return c.sha
}

func (c *CheckEvent) GetBranch() string {
	return c.branch
}

func (c *CheckEvent) GetToken() string {
	return c.token
}

func (c *CheckEvent) GetPRURL() string {
	return c.prURL
}

func (c *CheckEvent) GetGhClient() *github.Client {
	return c.ghClient
}

func NewCheckEvent(clientCreator githubapp.ClientCreator, event GenericGithubEvent, config *config.Config) (*CheckEvent, error) {
	repo := event.GetRepo()

	if ok, err := repo.IsValid(config); !ok {
		return nil, err
	}

	installationID := githubapp.GetInstallationIDFromEvent(event)
	client, err := clientCreator.NewAppClient()
	if err != nil {
		log.Error().Err(err).Msg("there was a problem while instantiating github client.")
		return nil, err
	}

	token, _, err := client.Apps.CreateInstallationToken(context.TODO(),
		installationID,
		&github.InstallationTokenOptions{
			RepositoryIDs: []int64{repo.GetID()},
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("there was a problem while creating installation token.")
		return nil, err
	}

	client, err = clientCreator.NewInstallationClient(githubapp.GetInstallationIDFromEvent(event))
	if err != nil {
		log.Error().Err(err).Msg("there was a problem while creating installation client.")
		return nil, err
	}

	return &CheckEvent{
		GenericGithubEvent:   event,
		repo:                 repo,
		sha:                  event.GetHeadSHA(),
		token:                token.GetToken(),
		branch:               event.GetHeadBranch(),
		ghClient:             client,
		prURL:                event.PrURL(),
		subFolderParallelism: config.SubFolderParallelism,
	}, nil
}

// GenericGithubEvent aims to factorize code for even treatment.
type GenericGithubEvent interface {
	githubapp.InstallationSource

	GetRepo() Repo
	GetHeadSHA() string
	GetHeadBranch() string
	IsValid() bool
	PrURL() string
	ExternalID() string
}

// Rename external struct to be able to extend them with interface func.
type CheckSuiteEvent struct{ *github.CheckSuiteEvent }

type (
	CheckRunEvent    struct{ *github.CheckRunEvent }
	PullRequestEvent struct{ *github.PullRequestEvent }
)

// CheckSuiteEvent.
func (e CheckSuiteEvent) GetRepo() Repo {
	return Repo{e.Repo}
}

func (e CheckSuiteEvent) GetHeadSHA() string {
	return e.GetCheckSuite().GetHeadSHA()
}

func (e CheckSuiteEvent) GetHeadBranch() string {
	return e.GetCheckSuite().GetHeadBranch()
}

func (e CheckSuiteEvent) IsValid() bool {
	if !utils.StrInSlice(getAuthorizedCheckSuiteActions(), e.GetAction()) {
		log.Debug().Msgf("Discarding event check_suite %s", e.GetAction())
		return false
	}

	if len(e.GetCheckSuite().PullRequests) == 0 {
		log.Debug().Msgf("Discarding event (not related to a PR)")
		return false
	}
	return true
}

func (e CheckSuiteEvent) PrURL() string {
	if prs := e.GetCheckSuite().PullRequests; len(prs) == 1 {
		return prs[0].GetHTMLURL()
	}
	return ""
}

func (e CheckSuiteEvent) ExternalID() string {
	return ""
}

// CheckRunEvent.
func (e CheckRunEvent) GetRepo() Repo {
	return Repo{e.Repo}
}

func (e CheckRunEvent) GetHeadSHA() string {
	return e.GetCheckRun().GetHeadSHA()
}

func (e CheckRunEvent) GetHeadBranch() string {
	return e.GetCheckRun().GetCheckSuite().GetHeadBranch()
}

func (e CheckRunEvent) IsValid() bool {
	if !utils.StrInSlice(getAuthorizedCheckRunActions(), e.GetAction()) {
		log.Debug().Msgf("Discarding event check_suite %s", e.GetAction())
		return false
	}
	if len(e.GetCheckRun().PullRequests) == 0 {
		log.Debug().Msgf("Discarding event (not related to a PR)")
		return false
	}
	return true
}

func (e CheckRunEvent) PrURL() string {
	if prs := e.GetCheckRun().PullRequests; len(prs) == 1 {
		return prs[0].GetHTMLURL()
	}
	return ""
}

func (e CheckRunEvent) ExternalID() string {
	return e.GetCheckRun().GetExternalID()
}

// PullRequestEvent.
func (e PullRequestEvent) GetRepo() Repo {
	return Repo{e.Repo}
}

func (e PullRequestEvent) GetHeadSHA() string {
	return e.GetPullRequest().GetHead().GetSHA()
}

func (e PullRequestEvent) GetHeadBranch() string {
	return e.GetPullRequest().GetHead().GetRef()
}

func (e PullRequestEvent) IsValid() bool {
	if !utils.StrInSlice(getAuthorizedPullRequestActions(), e.GetAction()) {
		log.Debug().Msgf("Discarding event pull_request %s", e.GetAction())
		return false
	}
	return true
}

func (e PullRequestEvent) PrURL() string {
	return e.GetPullRequest().GetURL()
}

func (e PullRequestEvent) ExternalID() string {
	return ""
}
