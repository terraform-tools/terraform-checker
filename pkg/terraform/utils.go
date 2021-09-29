package terraform

import (
	"io/fs"
	"path/filepath"
	"regexp"

	"github.com/rs/zerolog/log"
)

// FindAllTfDir finds all of the terraform directory inside a directory.
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
