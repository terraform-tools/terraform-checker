package errors

import (
	"errors"
	"fmt"
)

func RepoNotValidError(msg string) error {
	return fmt.Errorf("%w : %s", errors.New("repo not valid"), msg)
}

func ConfigNotValidError(msg string) error {
	return fmt.Errorf("%w : %s", errors.New("config not valid"), msg)
}
