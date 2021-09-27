package github

import "github.com/google/go-github/v39/github"

type Repo struct {
	Name     string
	FullName string
	Owner    string
	ID       int64
}

type GhCheckRun struct {
	Name string
	ID   int64
}

type CheckEvent struct {
	Repo       Repo
	Sha        string
	Token      string
	HeadBranch string
	GhClient   *github.Client
}
