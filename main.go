package main

import (
	cli "github.com/crane-cloud/mira-new/cmd/cli"
)

func main() {
	cli.Version = "0.1.0"
	cli.Execute()
}
