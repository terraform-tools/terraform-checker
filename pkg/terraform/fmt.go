package terraform

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/hashicorp/terraform-exec/tfexec"
)

func FixFmt(cloneDir string) error {
	for _, tfDir := range FindAllTfDir(cloneDir) {
		log.Info().Msgf("Executing action fmt on tfDir: %s", tfDir)
		tf, err := tfexec.NewTerraform(tfDir, terraformPath)
		if err != nil {
			return err
		}

		if err := tf.FormatWrite(context.TODO()); err != nil {
			return err
		}
	}
	return nil
}
