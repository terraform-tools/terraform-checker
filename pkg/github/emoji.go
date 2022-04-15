package github

import "github.com/shurcooL/githubv4"

// https://github.com/ikatyang/emoji-cheat-sheet/blob/master/README.md#symbols
func CheckConclusionStateEmoji(c githubv4.CheckConclusionState) string {
	stateToEmoji := map[githubv4.CheckConclusionState]string{
		githubv4.CheckConclusionStateActionRequired: ":question:",
		githubv4.CheckConclusionStateTimedOut:       ":hourglass:",
		githubv4.CheckConclusionStateCancelled:      "",
		githubv4.CheckConclusionStateFailure:        ":x:",
		githubv4.CheckConclusionStateSuccess:        ":heavy_check_mark:",
		githubv4.CheckConclusionStateNeutral:        "",
		githubv4.CheckConclusionStateSkipped:        ":next_track_button:",
		githubv4.CheckConclusionStateStartupFailure: ":x:",
		githubv4.CheckConclusionStateStale:          "",
	}
	return stateToEmoji[c]
}
