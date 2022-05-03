package terraform_test

import (
	"path"
	"path/filepath"
	"testing"

	"github.com/terraform-tools/terraform-checker/pkg/terraform"
)

func TestCheckTfFmt(t *testing.T) {
	t.Parallel()
	testDir, _ := filepath.Abs("../../test")

	testCases := []struct {
		directory string
		output    bool
	}{
		{
			directory: "terraform_ok",
			output:    true,
		}, {
			directory: "terraform_invalid",
			output:    true,
		}, {
			directory: "terraform_bad_fmt",
			output:    false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.directory, func(t *testing.T) {
			t.Parallel()
			ok, msg := terraform.CheckTfFmt(path.Join(testDir, tc.directory))
			if ok != tc.output {
				t.Errorf("CheckTfDir failed for dir %v, expected %v, got %v, message %v", tc.directory, tc.output, ok, msg)
			}
		})
	}
}

func TestCheckTfValidate(t *testing.T) {
	t.Parallel()
	testDir, _ := filepath.Abs("../../test")

	testCases := []struct {
		directory string
		output    bool
	}{
		{
			directory: "terraform_ok",
			output:    true,
		}, {
			directory: "terraform_invalid",
			output:    false,
		}, {
			directory: "terraform_bad_fmt",
			output:    true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.directory, func(t *testing.T) {
			t.Parallel()
			ok, msg, _ := terraform.CheckTfValidate(path.Join(testDir, tc.directory))
			if ok != tc.output {
				t.Errorf("CheckTfDir failed for dir %v, expected %v, got %v, message %v", tc.directory, tc.output, ok, msg)
			}
		})
	}
}
