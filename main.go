package main

import (
	cli "mira/cmd/cli"
)

func main() {
	cli.Version = "0.1.0"
	cli.Execute()
}
