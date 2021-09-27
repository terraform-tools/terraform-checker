package terraform

import (
	"context"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/rs/zerolog/log"
)

const terraformPath = "terraform"

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

func CheckTfDir(dir string) (bool, string) {
	workingDir := dir
	log.Info().Msg("Starting new Terraform")
	tf, err := tfexec.NewTerraform(workingDir, terraformPath)
	if err != nil {
		log.Error().Err(err).Msg("error creating Terraform object")
		return false, ""
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Error().Err(err).Msg("error running terraform init")
		return false, err.Error()
	}

	ok, output := tfValidate(tf)
	if !ok {
		return false, output
	}

	ok, output = tfFormat(tf)
	if !ok {
		return false, output
	}

	return true, ""
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
		return false, "The command `terraform validate` is failing on your repo with the following errors\n" + errorString
	}
	return true, ""
}

func tfFormat(tf *tfexec.Terraform) (bool, string) {
	ok, files, err := tf.FormatCheck(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("error running terraform fmt check")
		return false, ""
	}
	if !ok {
		log.Info().Msg("Running fmt")
		err := tf.FormatWrite(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("error running terraform fmt write")
			return false, "Error running `terraform fmt` " + err.Error()
		}
		return false, "Your terraform formatting is wrong for the following files:\n" + strings.Join(
			files,
			"\n",
		) + "\nplease run `terraform fmt` in the right dir or launch the `fmt` action" + "\n\n" + "more info [here](https://www.terraform.io/docs/cli/commands/fmt.html)"
	}
	return true, ""
}
