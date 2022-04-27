package terraform

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v43/github"
	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
	"github.com/terraform-linters/tflint/tflint"
)

// FindAllTfDir finds all of the terraform directory inside a directory.
func FindAllTfDir(dir string) (out []string) {
	regex := regexp.MustCompile("terraform.*")
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			} else if regex.MatchString(d.Name()) {
				out = append(out, path)
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("error walking dir")
	}

	return
}

// InitTfLint goal is to launch tflint --init once at program startup.
func InitTfLint() {
	ok, out := tfLintInit()
	if !ok {
		log.Error().Msgf("error while executing tflint --init. out: %s", out)
	}
	return
}

// TfLintRuleSeverityToAnnotationLevel allows to convert tflint severity to github annotation level.
func TfLintRuleSeverityToAnnotationLevel(severity string) *string {
	var finalStr githubv4.CheckAnnotationLevel

	switch severity {
	case tflint.ERROR.String():
		finalStr = githubv4.CheckAnnotationLevelFailure
	case tflint.NOTICE.String():
		finalStr = githubv4.CheckAnnotationLevelNotice
	case tflint.WARNING.String():
		finalStr = githubv4.CheckAnnotationLevelWarning
	default:
		finalStr = githubv4.CheckAnnotationLevelWarning
	}

	return github.String(strings.ToLower(string(finalStr)))
}
