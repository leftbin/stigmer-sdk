// Package examples demonstrates parallel execution using FORK tasks.
package main

import (
	"log"

	stigmeragent "github.com/leftbin/stigmer-sdk/go"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

// This example demonstrates parallel task execution using FORK.
//
// The workflow:
// 1. Fetches data that needs parallel processing
// 2. Forks into multiple parallel branches
// 3. Each branch processes data independently
// 4. Joins results and aggregates
func main() {
	defer stigmeragent.Complete()

	wf, err := workflow.New(
		workflow.WithNamespace("data-processing"),
		workflow.WithName("parallel-processing"),
		workflow.WithVersion("1.0.0"),
		workflow.WithDescription("Process data in parallel branches"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Task 1: Fetch data to process using high-level helper
	wf.AddTask(workflow.HttpCallTask("fetchData",
		workflow.WithHTTPGet(), // Type-safe HTTP method
		workflow.WithURI("https://api.example.com/data"),
	).ExportAll())

	// Task 2: Fork into parallel branches using field references and type-safe setters
	wf.AddTask(workflow.ForkTask("parallelProcessing",
		// Branch 1: Process analytics
		workflow.WithBranch("analytics",
			workflow.HttpCallTask("computeAnalytics",
				workflow.WithHTTPPost(), // Type-safe HTTP method
				workflow.WithURI("https://api.example.com/analytics"),
				workflow.WithBody(map[string]any{
					"data": workflow.FieldRef("data"),
					"type": "analytics",
				}),
			).ExportField("analytics"),

			workflow.SetTask("storeAnalytics",
				workflow.SetBool("analyticsComplete", true),
			),
		),

		// Branch 2: Process validation
		workflow.WithBranch("validation",
			workflow.HttpCallTask("validateData",
				workflow.WithHTTPPost(), // Type-safe HTTP method
				workflow.WithURI("https://api.example.com/validate"),
				workflow.WithBody(map[string]any{
					"data": workflow.FieldRef("data"),
				}),
			).ExportField("validationResult"),

			workflow.SetTask("storeValidation",
				workflow.SetBool("validationComplete", true),
			),
		),

		// Branch 3: Process transformation
		workflow.WithBranch("transformation",
			workflow.HttpCallTask("transformData",
				workflow.WithHTTPPost(), // Type-safe HTTP method
				workflow.WithURI("https://api.example.com/transform"),
				workflow.WithBody(map[string]any{
					"data":   workflow.FieldRef("data"),
					"format": "json",
				}),
			).ExportField("transformed"),

			workflow.SetTask("storeTransformed",
				workflow.SetBool("transformationComplete", true),
			),
		),

		// Branch 4: Process notification
		workflow.WithBranch("notification",
			workflow.HttpCallTask("sendNotification",
				workflow.WithHTTPPost(), // Type-safe HTTP method
				workflow.WithURI("https://api.example.com/notify"),
				workflow.WithBody(map[string]any{
					"message": "Processing started",
					"dataId":  workflow.FieldRef("data.id"),
				}),
			),

			workflow.SetTask("markNotified",
				workflow.SetBool("notificationSent", true),
			),
		),
	))

	// Task 3: Aggregate results (after all branches complete) using variable references
	wf.AddTask(workflow.SetTask("aggregateResults",
		workflow.SetString("status", "completed"),
		workflow.SetVar("analyticsStatus", workflow.VarRef("analyticsComplete")),
		workflow.SetVar("validationStatus", workflow.VarRef("validationComplete")),
		workflow.SetVar("transformationStatus", workflow.VarRef("transformationComplete")),
		workflow.SetVar("notificationStatus", workflow.VarRef("notificationSent")),
	))

	// Task 4: Send completion notification using variable references
	wf.AddTask(workflow.HttpCallTask("sendCompletion",
		workflow.WithHTTPPost(), // Type-safe HTTP method
		workflow.WithURI("https://api.example.com/completion"),
		workflow.WithBody(map[string]any{
			"status": workflow.VarRef("status"),
			"results": map[string]any{
				"analytics":      workflow.VarRef("analytics"),
				"validation":     workflow.VarRef("validationResult"),
				"transformation": workflow.VarRef("transformed"),
			},
		}),
	))

	log.Printf("Created parallel workflow: %s", wf)
}
