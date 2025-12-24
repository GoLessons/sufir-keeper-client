package main

import (
	"os"

	"github.com/GoLessons/sufir-keeper-client/internal/buildinfo"
	"github.com/GoLessons/sufir-keeper-client/internal/cli"
)

func main() {
	buildinfo.Ensure()
	version, commit, date := buildinfo.Info()

	app := cli.NewRootCmd(version, commit, date)
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}
