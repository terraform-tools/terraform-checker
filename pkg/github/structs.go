package github

import "github.com/google/go-github/v39/github"

type GithubRepo struct {
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
	Repo       GithubRepo
	Sha        string
	Token      string
	HeadBranch string
	GhClient   *github.Client
}
