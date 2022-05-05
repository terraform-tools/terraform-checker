package github

import (
	"github.com/terraform-tools/terraform-checker/pkg/git"
	"github.com/terraform-tools/terraform-checker/pkg/terraform"
)

const fmtCommitName = "terraform-checker fmt fix"

func (e *CheckEvent) fixFmt() error {
	repo, dir, err := git.CloneRepo(e.GetRepo().GetFullName(), e.GetSHA(), e.GetBranch(), e.GetToken())
	if err != nil {
		return err
	}

	if err := terraform.FixFmt(dir); err != nil {
		return err
	}

	return git.CommitAndPushRepo(fmtCommitName, repo)
}
