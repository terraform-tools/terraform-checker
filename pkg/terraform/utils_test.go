package terraform_test

import (
	"path/filepath"
	"testing"

	"github.com/terraform-tools/terraform-checker/pkg/terraform"
)

func TestFindAllTFDir(t *testing.T) {
	t.Parallel()

	path, err := filepath.Abs("../../test")
	if err != nil {
		t.Errorf("Error getting directory %v", err)
	}
	dirs := terraform.FindAllTfDir(path)

	if len(dirs) != 3 {
		t.Errorf("Expected to find 3 tfdir, got %v, %v %v", len(dirs), dirs, path)
	}
}
