package terraform

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-github/v43/github"
	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
	"github.com/terraform-linters/tflint/formatter"
)

// TfCheckType defines the different possible terraform checks.
type TfCheckType int64

const (
	Fmt TfCheckType = iota
	Validate
	TFLint
)

func (t TfCheckType) String() string {
	switch t {
	case Fmt:
		return "fmt"
	case Validate:
		return "validate"
	case TFLint:
		return "tflint"
	default:
		return "not implemented"
	}
}

func TfCheckTypeFromString(s string) TfCheckType {
	switch s {
	case "fmt":
		return Fmt
	case "validate":
		return Validate
	case "tflint":
		return TFLint
	default:
		return -1
	}
}

// TfCheck interface defines all functions that should be present for any TfCheck.
type TfCheck interface {
	Name() string
	Type() TfCheckType
	Run() (bool, string)
	Dir() string
	RelDir() string
	FailureConclusion() githubv4.CheckConclusionState
	FixActions() []*github.CheckRunAction
	Annotations() []*github.CheckRunAnnotation
}

type TfCheckFields struct {
	dir     string
	relDir  string
	checkOk bool
}

func NewTfCheckFields(dir, relDir string) TfCheckFields {
	return TfCheckFields{
		dir:    dir,
		relDir: relDir,
	}
}

func (t *TfCheckFields) Dir() string {
	return t.dir
}

func (t *TfCheckFields) RelDir() string {
	return t.relDir
}

// Fmt

type TfCheckFmt struct {
	TfCheckFields
}

func NewTfCheckFmt(tfDir, relDir string) *TfCheckFmt {
	return &TfCheckFmt{
		NewTfCheckFields(tfDir, relDir),
	}
}

func (t *TfCheckFmt) Name() string {
	return Fmt.String()
}

func (t *TfCheckFmt) Type() TfCheckType {
	return Fmt
}

func (t *TfCheckFmt) Run() (bool, string) {
	return CheckTfFmt(t.dir)
}

func (t *TfCheckFmt) FailureConclusion() githubv4.CheckConclusionState {
	return githubv4.CheckConclusionStateFailure
}

func (t *TfCheckFmt) FixActions() (actions []*github.CheckRunAction) {
	actions = append(actions, &github.CheckRunAction{
		// Max length 20 characters
		Label: "Trigger tf fmt",
		// Max length 40 characters
		Description: "Add a terraform fmt commit",
		// Max length 20 characters
		Identifier: t.Name(),
	})
	return
}

func (t *TfCheckFmt) Annotations() (annotations []*github.CheckRunAnnotation) {
	return
}

// Validate

type TfCheckValidate struct {
	TfCheckFields
}

func NewTfCheckValidate(tfDir, relDir string) *TfCheckValidate {
	return &TfCheckValidate{
		NewTfCheckFields(tfDir, relDir),
	}
}

func (t *TfCheckValidate) Name() string {
	return Validate.String()
}

func (t *TfCheckValidate) Type() TfCheckType {
	return Validate
}

func (t *TfCheckValidate) Run() (bool, string) {
	return CheckTfValidate(t.dir)
}

func (t *TfCheckValidate) FailureConclusion() githubv4.CheckConclusionState {
	return githubv4.CheckConclusionStateFailure
}

func (t *TfCheckValidate) FixActions() (actions []*github.CheckRunAction) {
	return
}

func (t *TfCheckValidate) Annotations() (annotations []*github.CheckRunAnnotation) {
	return
}

// TFLint

type TfCheckTfLint struct {
	TfCheckFields
	tfLintOutput formatter.JSONOutput
}

func NewTfCheckTfLint(tfDir, relDir string) *TfCheckTfLint {
	return &TfCheckTfLint{
		TfCheckFields: NewTfCheckFields(tfDir, relDir),
	}
}

func (t *TfCheckTfLint) Name() string {
	return TFLint.String()
}

func (t *TfCheckTfLint) Type() TfCheckType {
	return TFLint
}

func (t *TfCheckTfLint) Run() (bool, string) {
	ok, out := tfLint(t.dir, "default")
	_, outJSONStr := tfLint(t.dir, "json")

	var outJSON formatter.JSONOutput
	if err := json.Unmarshal([]byte(outJSONStr), &outJSON); err != nil {
		log.Error().Err(err).Msg("error unmarshalling tflint output")
		return false, out
	}
	t.tfLintOutput = outJSON

	return ok, out
}

func (t *TfCheckTfLint) FailureConclusion() githubv4.CheckConclusionState {
	return githubv4.CheckConclusionStateFailure
}

func (t *TfCheckTfLint) FixActions() (actions []*github.CheckRunAction) {
	return
}

func (t *TfCheckTfLint) Annotations() (annotations []*github.CheckRunAnnotation) {
	for _, issue := range t.tfLintOutput.Issues {
		currentIssue := issue

		if issue.Range.Filename == "" {
			continue
		}

		newAnnotation := github.CheckRunAnnotation{
			Title:           github.String(currentIssue.Rule.Name),
			Message:         &currentIssue.Message,
			Path:            github.String(fmt.Sprintf("%s/%s", t.RelDir(), currentIssue.Range.Filename)),
			AnnotationLevel: github.String(strings.ToLower(TfLintRuleSeverityToAnnotationLevel(currentIssue.Rule.Severity))),
		}

		// Only set StarLine/EndLine if they are different from 0
		if currentIssue.Range.Start.Line == 0 && currentIssue.Range.End.Line == 0 {
			continue
		}

		newAnnotation.StartLine = github.Int(currentIssue.Range.Start.Line)
		newAnnotation.EndLine = github.Int(currentIssue.Range.End.Line)

		// Only set StarColumn/EndColumn if StartLine/Endline are on same line
		if newAnnotation.StartLine == newAnnotation.EndLine {
			newAnnotation.StartColumn = github.Int(currentIssue.Range.Start.Column)
			newAnnotation.EndColumn = github.Int(currentIssue.Range.End.Column)
		}

		annotations = append(annotations, &newAnnotation)
	}

	return annotations
}

func NewTfCheck(checkType TfCheckType, tfDir, relDir string) TfCheck {
	switch checkType {
	case Fmt:
		return NewTfCheckFmt(tfDir, relDir)
	case Validate:
		return NewTfCheckValidate(tfDir, relDir)
	case TFLint:
		return NewTfCheckTfLint(tfDir, relDir)
	default:
		return nil
	}
}

func GetAllTfChecks(tfDir, relDir string) (checks []TfCheck) {
	return []TfCheck{
		NewTfCheckFmt(tfDir, relDir),
		NewTfCheckValidate(tfDir, relDir),
		NewTfCheckTfLint(tfDir, relDir),
	}
}

func GetTfChecks(tfDir, relDir string, checkTypes []string) (checks []TfCheck) {
	if len(checkTypes) > 0 {
		for _, c := range checkTypes {
			checks = append(checks, NewTfCheck(TfCheckTypeFromString(c), tfDir, relDir))
		}
		return
	}

	checks = append(checks, GetAllTfChecks(tfDir, relDir)...)
	return
}
