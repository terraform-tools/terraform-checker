package git

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func CloneRepo(repoName string, hash string, headBranch string, ghToken string) (*git.Repository, string, error) {
	dir, err := ioutil.TempDir("", "tf-checker")
	if err != nil {
		return nil, "", err
	}

	log.Debug().Msgf("Cloning repo %s into %s ...", repoName, dir)
	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: "https://x-access-token:" + ghToken + "@github.com/" + repoName,
	})
	if err != nil {
		return nil, "", err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return nil, "", err
	}
	err = wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(hash)})
	if err != nil {
		return nil, "", err
	}
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(headBranch),
		Create: true,
	})

	return repo, dir, err
}

func CommitAndPushRepo(commitMsg string, repo *git.Repository) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	_, err = w.Add(".")
	if err != nil {
		return err
	}

	status, err := w.Status()
	if status.IsClean() {
		log.Debug().Err(err).Msg("Directory is clean, not committing")
		return nil
	}

	_, err = w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "terraform-checker",
			Email: "terraform-checker@terraform-checker.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Error committing")
		return err
	}
	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
	})
	if err != nil {
		log.Error().Err(err).Msg("Error pushing")
	}
	return err
}

func RemoveRepo(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		log.Error().Err(err).Msgf("Error while removing folder %v", dir)
	}
}
