package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/rs/zerolog/log"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/formatter"
)

const terraformPath = "terraform"

func CheckTfFmt(dir string) (bool, string) {
	ok, output, tf := tfInit(dir)
	if !ok {
		return ok, output
	}

	return tfFormat(tf)
}

func CheckTfValidate(dir string) (bool, string, *tfjson.ValidateOutput) {
	if ok, output, _ := tfInit(dir); !ok {
		return ok, output, nil
	}

	ok, output := tfValidate(dir, false)
	_, outputJSON := tfValidate(dir, true)

	var outJSON tfjson.ValidateOutput
	if err := json.Unmarshal([]byte(outputJSON), &outJSON); err != nil {
		log.Error().Err(err).Msg("error unmarshalling terraform validate output")
		return false, output, nil
	}

	return ok, output, &outJSON
}

func CheckTfLint(dir string) (bool, string, *formatter.JSONOutput) {
	ok, output, _ := tfInit(dir)
	if !ok {
		return ok, output, nil
	}

	ok, out := tfLint(dir, "default")
	_, outJSONStr := tfLint(dir, "json")

	var outJSON formatter.JSONOutput
	if err := json.Unmarshal([]byte(outJSONStr), &outJSON); err != nil {
		log.Error().Err(err).Msg("error unmarshalling tflint output")
		return false, out, nil
	}

	return tfLintStatus(&outJSON), out, &outJSON
}

func tfInit(dir string) (bool, string, *tfexec.Terraform) {
	workingDir := dir
	tf, err := tfexec.NewTerraform(workingDir, terraformPath)
	if err != nil {
		log.Error().Err(err).Msg("error creating Terraform object")
		return false, "", nil
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true), tfexec.Backend(false))
	if err != nil {
		log.Error().Err(err).Msg("error running terraform init")
		return false, err.Error(), nil
	}

	return true, "", tf
}

func tfValidate(dir string, json bool) (bool, string) {
	args := []string{"validate", "-no-color"}
	if json {
		args = append(args, "-json")
	}
	cmd := exec.Command("terraform", args...) // #nosec
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return err == nil, string(out)
}

func tfFormat(tf *tfexec.Terraform) (bool, string) {
	ok, files, err := tf.FormatCheck(context.Background(), &tfexec.RecursiveOption{})
	if err != nil {
		log.Error().Err(err).Msg("error running terraform fmt check")
		return false, ""
	}
	if !ok {
		return false, "Your terraform formatting is wrong for the following files:\n" + strings.Join(
			files,
			"\n",
		) + "\nplease run `terraform fmt -recursive` in the right dir or launch the `Trigger tf fmt` action ⬆️⬆️⬆️" + "\n\n" + "more info [here](https://www.terraform.io/docs/cli/commands/fmt.html)"
	}
	return true, ""
}

func tfLint(dir, format string) (bool, string) {
	cmd := exec.Command("tflint", []string{fmt.Sprintf("-f=%s", format)}...) // #nosec
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return err == nil, string(out)
}

func tfLintInit() (bool, string) {
	cmd := exec.Command("tflint", []string{"--init"}...) // #nosec
	out, err := cmd.CombinedOutput()
	return err == nil, string(out)
}

func tfLintStatus(out *formatter.JSONOutput) bool {
	if len(out.Errors) > 0 {
		return false
	}
	for _, i := range out.Issues {
		if i.Rule.Severity == tflint.ERROR.String() {
			return false
		}
	}
	return true
}
