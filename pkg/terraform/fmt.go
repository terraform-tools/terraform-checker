package terraform

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/hashicorp/terraform-exec/tfexec"
)

func FixFmt(dir string) error {
	for _, tfDir := range FindAllTfDir(dir) {
		log.Print("fixing ", tfDir)
		tf, err := tfexec.NewTerraform(tfDir, terraformPath)
		if err != nil {
			return err
		}

		if err := tf.FormatWrite(context.TODO()); err != nil {
			return err
		}
	}
	log.Print("pushing commit")
	return nil
}
