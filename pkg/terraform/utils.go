package terraform

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v43/github"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
	"github.com/terraform-linters/tflint/tflint"
	"github.com/terraform-tools/terraform-checker/pkg/utils"
)

// FindAllTfDir finds all of the terraform directory inside a directory.
func FindAllTfDir(dir string) (out []string) {
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == ".terraform" {
				return filepath.SkipDir
			}
		}

		currentPath := filepath.Dir(path)
		if strings.HasSuffix(path, ".tf") && !utils.StrInSlice(out, currentPath) {
			out = append(out, currentPath)
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

// TfValidateSeverityToAnnotationLevel allows to convert terraform validate severity to github annotation level.
func TfValidateSeverityToAnnotationLevel(severity tfjson.DiagnosticSeverity) *string {
	var finalStr githubv4.CheckAnnotationLevel

	switch severity {
	case tfjson.DiagnosticSeverityUnknown:
		finalStr = githubv4.CheckAnnotationLevelFailure
	case tfjson.DiagnosticSeverityError:
		finalStr = githubv4.CheckAnnotationLevelFailure
	case tfjson.DiagnosticSeverityWarning:
		finalStr = githubv4.CheckAnnotationLevelWarning
	default:
		finalStr = githubv4.CheckAnnotationLevelWarning
	}

	return github.String(strings.ToLower(string(finalStr)))
}
