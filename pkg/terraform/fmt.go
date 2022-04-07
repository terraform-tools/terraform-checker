package terraform

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/terraform-tools/terraform-checker/pkg/filter"

	"github.com/hashicorp/terraform-exec/tfexec"
)

func FixFmt(cloneDir string, filters ...filter.Option) error {
	dirFilter := ""
	for _, f := range filters {
		switch f := f.(type) {
		case *filter.DirFilter:
			if f.Dir != "" {
				dirFilter = f.Dir
			}
		default:
		}
	}

	for _, tfDir := range FindAllTfDir(cloneDir) {
		// If dirFilter is defined and current dir does not match, continue
		if dirFilter != "" && !strings.Contains(tfDir, dirFilter) {
			continue
		}

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
