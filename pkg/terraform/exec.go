package terraform

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/rs/zerolog/log"
)

const terraformPath = "terraform"

func CheckTfFmt(dir string) (bool, string) {
	ok, output, tf := tfInit(dir)
	if !ok {
		return ok, output
	}

	return tfFormat(tf)
}

func CheckTfValidate(dir string) (bool, string) {
	ok, output, tf := tfInit(dir)
	if !ok {
		return ok, output
	}

	return tfValidate(tf)
}

func CheckTfLint(dir string) (bool, string) {
	ok, output, tf := tfInit(dir)
	if !ok {
		return ok, output
	}

	return tfValidate(tf)
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

func tfValidate(tf *tfexec.Terraform) (bool, string) {
	validationError, err := tf.Validate(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("error running terraform validate")
		return false, "Error running `terraform validate` " + err.Error()
	}
	if validationError.ErrorCount != 0 {
		errorString := ""
		for _, diag := range validationError.Diagnostics {
			errorString += "`" + diag.Range.Filename + "` line:`" + strconv.Itoa(diag.Range.Start.Line) + "`"
			errorString += "\n"
			errorString += diag.Summary
			errorString += "\n"
			errorString += diag.Detail
			errorString += "\n"
			errorString += "\n"
		}
		errorString += "\n"
		return false, ("The command `terraform validate` is failing on your repo with the following errors\n" + errorString +
			"\nYou can reproduce locally by executing `terraform init -backend=false;terraform validate`")
	}
	return true, ""
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
