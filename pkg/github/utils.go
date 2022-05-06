package github

func getAuthorizedCheckSuiteActions() []string {
	return []string{"requested", "rerequested"}
}

func getAuthorizedCheckRunActions() []string {
	return []string{"requested_action", "rerequested"}
}

func getAuthorizedPullRequestActions() []string {
	return []string{"opened", "reopened"}
}
