package main

import (
	"errors"
	"fmt"
	"os"

	stepflow "github.com/f24aalam/stepflow/pkg"
)

func main() {
	wizard := stepflow.New().WithSteps(
		stepflow.Text("app", "Application name").Default("my-app"),
		stepflow.Confirm("docker", "Enable Docker containerization?").
			Default("Yes").
			WithNext(func(r stepflow.Result) []stepflow.Step {
				if r.Bool("docker") {
					// Dynamically append Docker configuration steps
					return []stepflow.Step{
						stepflow.List("runtime", "Select container base image").
							Items(
								stepflow.Item("alpine", "Lightweight musl (~5MB)"),
								stepflow.Item("debian-slim", "glibc compatibility"),
								stepflow.Item("distroless", "Minimal security-hardened"),
							).
							PreSelect("distroless"),
						stepflow.Text("port", "Exposed container port").
							Placeholder("e.g. 8080").
							Default("8080"),
					}
				}

				// Otherwise append native build note
				return []stepflow.Step{
					stepflow.Info("native_note", "Native Compilation").
						Body("Docker skipped. Building standard host binary."),
				}
			}),
	)

	result, err := wizard.Run()
	if err != nil {
		if errors.Is(err, stepflow.ErrCancelled) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("\nApp: %s\n", result.Get("app"))
	fmt.Printf("Docker enabled: %t\n", result.Bool("docker"))
	if result.Bool("docker") {
		fmt.Printf("Base image: %s, Port: %s\n", result.Get("runtime"), result.Get("port"))
	}
}
