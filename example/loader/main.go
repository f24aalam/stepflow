package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	stepflow "github.com/f24aalam/stepflow/pkg"
)

func main() {
	wizard := stepflow.New().WithSteps(
		stepflow.Text("repo", "Repository URL").Default("github.com/charmbracelet/bubbletea"),
		stepflow.Load("clone", "Cloning and scanning repository").
			RunContext(func(ctx context.Context, status chan<- string) (string, error) {
				tasks := []string{
					"Connecting to remote host…",
					"Receiving objects: 25%…",
					"Receiving objects: 75%…",
					"Receiving objects: 100% done.",
					"Resolving deltas…",
					"Scanning AST symbols…",
				}

				for _, t := range tasks {
					select {
					case <-ctx.Done():
						return "", ctx.Err()
					case status <- t:
					}
					time.Sleep(300 * time.Millisecond)
				}

				return "Repository cloned & analyzed", nil
			}),
		stepflow.Success("complete", "Ready").Body("Analysis complete! Press enter to finish."),
	)

	// Run with context so Ctrl+C or cancellation will terminate the worker promptly
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result, err := wizard.RunContext(ctx)
	if err != nil {
		if errors.Is(err, stepflow.ErrCancelled) || errors.Is(err, context.Canceled) {
			fmt.Println("\nOperation cancelled by user.")
			return
		}
		fmt.Fprintf(os.Stderr, "\nWorker error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nRepository: %s\n", result.Get("repo"))
	fmt.Printf("Worker Result: %s\n", result.Get("clone"))
}
