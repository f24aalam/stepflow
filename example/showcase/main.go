package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	stepflow "github.com/f24aalam/stepflow/pkg"
)

func main() {
	wizard := stepflow.New().WithSteps(
		// 1. Info Step: introductory overview
		stepflow.Info("welcome", "Welcome to Service Provisioner").
			Body("This wizard scaffolds a production-ready Go service.\nFollow the prompts to configure components and deployment."),

		// 2. Text Step: textinput with placeholder and default value
		stepflow.Text("name", "Service Name").
			Placeholder("e.g. auth-service").
			Default("orders-api"),

		// 3. Confirm Step: default "No" with up/down/k/j navigation
		stepflow.Confirm("deploy_cloud", "Deploy to cloud immediately?").
			Default("No"),

		// 4. List Step: multi-select with pinned items, visible row pagination, metadata, and search filtering
		stepflow.List("components", "Select components to bundle").
			MultiSelect(true).
			VisibleRows(4).
			Pinned("Core Stack", "Recommended defaults", "HTTP Router", "Logger").
			Items(
				stepflow.Item("HTTP Router", "Chi / standard net/http"),
				stepflow.Item("Logger", "Structured slog wrapper"),
				stepflow.Item("Postgres", "SQL database client & migrations"),
				stepflow.Item("Redis", "In-memory caching store"),
				stepflow.Item("OpenTelemetry", "Distributed tracing & metrics"),
				stepflow.Item("gRPC Server", "Protocol Buffers RPC interface"),
				stepflow.Item("GraphQL Engine", "Declarative query schema"),
				stepflow.Item("RabbitMQ", "AMQP asynchronous message queue"),
				stepflow.Item("Docker Setup", "Multi-stage Dockerfile and Compose"),
			).
			PreSelect("HTTP Router", "Logger").
			// 5. Dynamic Branching: dynamically inject steps based on previous answers
			WithNext(func(r stepflow.Result) []stepflow.Step {
				var dynamicSteps []stepflow.Step

				// If Postgres was selected in multi-select, ask for database configuration
				hasPostgres := false
				for _, c := range r.Strings("components") {
					if c == "Postgres" {
						hasPostgres = true
						break
					}
				}

				if hasPostgres {
					dynamicSteps = append(dynamicSteps,
						stepflow.Text("db_name", "PostgreSQL database name").
							Placeholder("e.g. production_db").
							Default("orders_db"),
						stepflow.Confirm("run_migrations", "Auto-run migrations on startup?").
							Default("Yes"),
					)
				}

				// If deploy_cloud was Yes, ask for cloud provider
				if r.Bool("deploy_cloud") {
					dynamicSteps = append(dynamicSteps,
						stepflow.List("cloud_provider", "Target cloud environment").
							Items(
								stepflow.Item("AWS", "ECS / Fargate"),
								stepflow.Item("GCP", "Cloud Run"),
								stepflow.Item("Azure", "Container Apps"),
							).
							PreSelect("GCP"),
					)
				}

				// 6. Load Step: async background worker with real-time status streaming
				dynamicSteps = append(dynamicSteps,
					stepflow.Load("scaffold", "Scaffolding project").
						RunContext(func(ctx context.Context, status chan<- string) (string, error) {
							stages := []string{
								"Creating project directories…",
								"Writing template files…",
								"Resolving Go module dependencies…",
								"Formatting and running tests…",
							}
							for _, stage := range stages {
								select {
								case <-ctx.Done():
									return "", ctx.Err()
								case status <- stage:
								}
								time.Sleep(350 * time.Millisecond)
							}
							return "Scaffolding completed", nil
						}),
					// 7. Terminal Step: Success screen
					stepflow.Success("summary", "Service Scaffolding Finished!").
						Body("All files generated and dependencies verified.\nRun 'cd "+r.Get("name")+" && go run .' to start."),
				)

				return dynamicSteps
			}),
	)

	result, err := wizard.Run()
	if err != nil {
		if errors.Is(err, stepflow.ErrCancelled) {
			fmt.Println("\nSetup cancelled.")
			return
		}
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}

	// Print summary of answers extracted using different Result helper methods
	fmt.Println("\n================ Configuration Summary ================")
	fmt.Printf("Service:        %s\n", result.Get("name"))
	fmt.Printf("Deploy Cloud:   %t\n", result.Bool("deploy_cloud"))
	if cloud := result.Get("cloud_provider"); cloud != "" {
		fmt.Printf("Cloud Provider: %s\n", cloud)
	}
	fmt.Printf("Components:     %s\n", strings.Join(result.Strings("components"), ", "))
	if db := result.Get("db_name"); db != "" {
		fmt.Printf("Database Name:  %s\n", db)
		fmt.Printf("Run Migrations: %t\n", result.Bool("run_migrations"))
	}
	fmt.Printf("Status:         %s\n", result.Get("scaffold"))
	fmt.Println("=======================================================")
}
