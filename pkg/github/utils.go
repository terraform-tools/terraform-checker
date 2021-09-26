package github

import (
	"context"

	"github.com/google/go-github/v39/github"
	"github.com/palantir/go-githubapp/githubapp"
)

func ghRepoFromCheckRunEvent(event github.CheckRunEvent) (GithubRepo, string, string) {
	repo := event.GetRepo()

	return GithubRepo{
		Name:     repo.GetName(),
		FullName: repo.GetFullName(),
		Owner:    repo.GetOwner().GetLogin(),
		ID:       repo.GetID(),
	}, event.GetCheckRun().GetHeadSHA(), event.GetCheckRun().GetCheckSuite().GetHeadBranch()
}

func ghRepoFromCheckSuiteEvent(event github.CheckSuiteEvent) (GithubRepo, string, string) {
	repo := event.GetRepo()

	return GithubRepo{
		Name:     repo.GetName(),
		FullName: repo.GetFullName(),
		Owner:    repo.GetOwner().GetLogin(),
		ID:       repo.GetID(),
	}, event.GetCheckSuite().GetHeadSHA(), event.GetCheckSuite().GetHeadBranch()
}

func (h *CheckHandler) getToken(
	ctx context.Context,
	event githubapp.InstallationSource,
	repo GithubRepo,
) (string, error) {
	installationID := githubapp.GetInstallationIDFromEvent(event)
	client, err := h.Client.NewAppClient()
	if err != nil {
		return "", err
	}

	token, _, err := client.Apps.CreateInstallationToken(ctx,
		installationID,
		&github.InstallationTokenOptions{
			RepositoryIDs: []int64{repo.ID},
		})

	return token.GetToken(), err
}

func (h *CheckHandler) getTokenAndClient(
	ctx context.Context,
	event githubapp.InstallationSource,
	repo GithubRepo,
) (token string, client *github.Client, err error) {
	token, err = h.getToken(ctx, event, repo)
	if err != nil {
		return
	}

	client, err = h.Client.NewInstallationClient(githubapp.GetInstallationIDFromEvent(event))
	return
}

func (h *CheckHandler) checkEventFromSuite(ctx context.Context, event github.CheckSuiteEvent) (*CheckEvent, error) {
	repo, sha, headBranch := ghRepoFromCheckSuiteEvent(event)

	token, client, err := h.getTokenAndClient(ctx, &event, repo)
	if err != nil {
		return nil, err
	}

	return &CheckEvent{
		Repo:       repo,
		Sha:        sha,
		Token:      token,
		HeadBranch: headBranch,
		GhClient:   client,
	}, nil
}

func (h *CheckHandler) checkEventFromRun(ctx context.Context, event github.CheckRunEvent) (*CheckEvent, error) {
	repo, sha, headBranch := ghRepoFromCheckRunEvent(event)

	token, client, err := h.getTokenAndClient(ctx, &event, repo)
	if err != nil {
		return nil, err
	}

	return &CheckEvent{
		Repo:       repo,
		Sha:        sha,
		Token:      token,
		HeadBranch: headBranch,
		GhClient:   client,
	}, nil
}
