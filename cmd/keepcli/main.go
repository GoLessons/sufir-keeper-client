package main

import "github.com/GoLessons/sufir-keeper-server/internal/cli"

func main() {
	app := cli.NewRootCmd()
	app.Execute()
}
