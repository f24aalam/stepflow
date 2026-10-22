package main

import (
	"errors"
	"fmt"
	stepflow "github.com/f24aalam/stepflow/pkg"
	"os"
)

func main() {
	result, err := stepflow.New().WithSteps(
		stepflow.Text("project", "Project name").Default("my-cli"),
		stepflow.Confirm("git", "Initialize Git?"),
		stepflow.List("license", "Choose license").Items(stepflow.Item("MIT"), stepflow.Item("Apache-2.0")).PreSelect("MIT"),
	).Run()
	if err != nil {
		if errors.Is(err, stepflow.ErrCancelled) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Project: %s, git: %t, license: %s\n", result.Get("project"), result.Bool("git"), result.Get("license"))
}
