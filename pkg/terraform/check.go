package terraform

import (
	"fmt"

	"github.com/google/go-github/v56/github"
	tfjson "github.com/hashicorp/terraform-json"
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

func AllTfCheckTypes() []string {
	return []string{
		Fmt.String(),
		Validate.String(),
		TFLint.String(),
	}
}

// TfCheck interface defines all functions that should be present for any TfCheck.
type TfCheck interface {
	Name() string
	Type() TfCheckType
	Run()
	Dir() string
	RelDir() string
	IsOK() bool
	Output() string
	FailureConclusion() githubv4.CheckConclusionState
	FixAction() *github.CheckRunAction
	Annotations() []*github.CheckRunAnnotation
}

type TfCheckFields struct {
	dir     string
	relDir  string
	checkOk bool
	output  string
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

func (t *TfCheckFields) IsOK() bool {
	return t.checkOk
}

func (t *TfCheckFields) Output() string {
	return t.output
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

func (t *TfCheckFmt) Run() {
	ok, out := CheckTfFmt(t.dir)
	t.checkOk = ok
	t.output = out
}

func (t *TfCheckFmt) FailureConclusion() githubv4.CheckConclusionState {
	return githubv4.CheckConclusionStateFailure
}

func (t *TfCheckFmt) FixAction() *github.CheckRunAction {
	return &github.CheckRunAction{
		// Max length 20 characters
		Label: "Trigger tf fmt",
		// Max length 40 characters
		Description: "Add a terraform fmt commit",
		// Max length 20 characters
		Identifier: t.Name(),
	}
}

func (t *TfCheckFmt) Annotations() (annotations []*github.CheckRunAnnotation) {
	return
}

// Validate

type TfCheckValidate struct {
	TfCheckFields
	tfValidateOutput *tfjson.ValidateOutput
}

func NewTfCheckValidate(tfDir, relDir string) *TfCheckValidate {
	return &TfCheckValidate{
		TfCheckFields: NewTfCheckFields(tfDir, relDir),
	}
}

func (t *TfCheckValidate) Name() string {
	return Validate.String()
}

func (t *TfCheckValidate) Type() TfCheckType {
	return Validate
}

func (t *TfCheckValidate) Run() {
	ok, out, tfValidateOutput := CheckTfValidate(t.dir)
	t.checkOk = ok
	t.output = out
	t.tfValidateOutput = tfValidateOutput
}

func (t *TfCheckValidate) FailureConclusion() githubv4.CheckConclusionState {
	return githubv4.CheckConclusionStateFailure
}

func (t *TfCheckValidate) FixAction() *github.CheckRunAction {
	return nil
}

func (t *TfCheckValidate) Annotations() (annotations []*github.CheckRunAnnotation) {
	if t.tfValidateOutput == nil || t.tfValidateOutput.Valid || t.tfValidateOutput.Diagnostics == nil {
		return annotations
	}

	for _, diag := range t.tfValidateOutput.Diagnostics {
		currentDiag := diag

		if currentDiag.Range == nil || currentDiag.Range.Filename == "" {
			continue
		}

		newAnnotation := github.CheckRunAnnotation{
			Title:           github.String(currentDiag.Summary),
			Message:         &currentDiag.Detail,
			Path:            github.String(fmt.Sprintf("%s/%s", t.RelDir(), currentDiag.Range.Filename)),
			AnnotationLevel: TfValidateSeverityToAnnotationLevel(currentDiag.Severity),
		}

		// Only set StarLine/EndLine if they are different from 0
		if currentDiag.Range.Start.Line == 0 && currentDiag.Range.End.Line == 0 {
			continue
		}

		newAnnotation.StartLine = github.Int(currentDiag.Range.Start.Line)
		newAnnotation.EndLine = github.Int(currentDiag.Range.End.Line)

		// Only set StarColumn/EndColumn if StartLine/Endline are on same line
		if newAnnotation.StartLine == newAnnotation.EndLine {
			newAnnotation.StartColumn = github.Int(currentDiag.Range.Start.Column)
			newAnnotation.EndColumn = github.Int(currentDiag.Range.End.Column)
		}

		annotations = append(annotations, &newAnnotation)
	}

	return annotations
}

// TFLint

type TfCheckTfLint struct {
	TfCheckFields
	tfLintOutput *formatter.JSONOutput
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

func (t *TfCheckTfLint) Run() {
	ok, out, tfLintOutput := CheckTfLint(t.dir)
	t.checkOk = ok
	t.output = out
	t.tfLintOutput = tfLintOutput
}

func (t *TfCheckTfLint) FailureConclusion() githubv4.CheckConclusionState {
	return githubv4.CheckConclusionStateFailure
}

func (t *TfCheckTfLint) FixAction() *github.CheckRunAction {
	return nil
}

func (t *TfCheckTfLint) Annotations() (annotations []*github.CheckRunAnnotation) {
	if t.tfLintOutput == nil {
		return annotations
	}
	for _, issue := range t.tfLintOutput.Issues {
		currentIssue := issue

		if issue.Range.Filename == "" {
			continue
		}

		newAnnotation := github.CheckRunAnnotation{
			Title:           github.String(currentIssue.Rule.Name),
			Message:         &currentIssue.Message,
			Path:            github.String(fmt.Sprintf("%s/%s", t.RelDir(), currentIssue.Range.Filename)),
			AnnotationLevel: TfLintRuleSeverityToAnnotationLevel(currentIssue.Rule.Severity),
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

func GetTfChecks(tfDir, relDir string, checkTypes []string) (checks []TfCheck) {
	for _, c := range checkTypes {
		checks = append(checks, NewTfCheck(TfCheckTypeFromString(c), tfDir, relDir))
	}
	return
}
