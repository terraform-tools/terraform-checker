package terraform

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v56/github"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-tools/terraform-checker/pkg/utils"
	"gopkg.in/yaml.v3"
)

const (
	tfDirConfigName     = ".tf-checker"
	tfDirEnabledDefault = true
)

type TfDir struct {
	path    string
	enabled bool
}

func (t *TfDir) Path() string {
	return t.path
}

func (t *TfDir) IsEnabled() bool {
	return t.enabled
}

type TfDirConfigFile struct {
	Enabled bool `yaml:"enabled"`
}

func parseTfDirConfig(path string) TfDirConfigFile {
	var data []byte
	var err error

	t := TfDirConfigFile{
		Enabled: tfDirEnabledDefault,
	}

	if data, err = os.ReadFile(path); err != nil {
		return t
	}

	if err = yaml.Unmarshal(data, &t); err != nil {
		log.Error().Err(err).Msgf("error while parsing tfDir config file %s", path)
	}
	return t
}

func NewTfDir(path string) *TfDir {
	newTfDir := TfDir{
		path: path,
	}
	conf := parseTfDirConfig(fmt.Sprintf("%s/%s", path, tfDirConfigName))
	newTfDir.enabled = conf.Enabled
	return &newTfDir
}

// FindAllTfDir finds all of the terraform directory inside a directory.
func FindAllTfDir(dir string) (tfDirs []*TfDir) {
	tfDirsPaths := []string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == ".terraform" {
				return filepath.SkipDir
			}
		}

		currentPath := filepath.Dir(path)
		if strings.HasSuffix(path, ".tf") && !utils.StrInSlice(tfDirsPaths, currentPath) {
			tfDirsPaths = append(tfDirsPaths, currentPath)
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("error walking dir")
	}

	for _, t := range tfDirsPaths {
		tfDirs = append(tfDirs, NewTfDir(t))
	}

	return
}

// InitTfLint goal is to launch tflint --init once at program startup.
func InitTfLint() {
	ok, out := tfLintInit()
	if !ok {
		log.Error().Msgf("error while executing tflint --init. out: %s", out)
	}
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
