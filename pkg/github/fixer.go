package github

import (
	"github.com/terraform-tools/terraform-checker/pkg/git"
	"github.com/terraform-tools/terraform-checker/pkg/terraform"
)

func (t *CheckEvent) fixFmt() error {
	repo, dir, err := git.CloneRepo(t.Repo.FullName, t.Sha, t.HeadBranch, t.Token)
	if err != nil {
		return err
	}

	if err := terraform.FixFmt(dir); err != nil {
		return err
	}

	return git.CommitAndPushRepo("test fmt", repo)
}
