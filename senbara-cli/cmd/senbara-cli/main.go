package main

import "github.com/pojntfx/senbara/senbara-cli/cmd/senbara-cli/cmd"

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
