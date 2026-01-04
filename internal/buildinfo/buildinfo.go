package buildinfo

import (
	"errors"
	"os"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func Info() (string, string, string) {
	return version, commit, date
}

func Validate() error {
	if version == "" || version == "dev" {
		return errors.New("version must be provided via -ldflags")
	}
	if date == "" || date == "unknown" {
		return errors.New("date must be provided via -ldflags")
	}
	return nil
}

func Ensure() {
	if err := Validate(); err != nil {
		os.Stderr.WriteString("build metadata warning: " + err.Error() + "\n")
	}
}
