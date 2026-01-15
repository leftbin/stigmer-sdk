// Package examples demonstrates parallel execution using FORK tasks.
package main

import (
	"log"

	"github.com/leftbin/stigmer-sdk/go/synthesis"
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
	defer synthesis.AutoSynth()

	wf, err := workflow.New(
		workflow.WithNamespace("data-processing"),
		workflow.WithName("parallel-processing"),
		workflow.WithVersion("1.0.0"),
		workflow.WithDescription("Process data in parallel branches"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Task 1: Fetch data to process
	wf.AddTask(workflow.HttpCallTask("fetchData",
		workflow.WithMethod("GET"),
		workflow.WithURI("https://api.example.com/data"),
	).Export("${.}"))

	// Task 2: Fork into parallel branches
	wf.AddTask(workflow.ForkTask("parallelProcessing",
		// Branch 1: Process analytics
		workflow.WithBranch("analytics",
			workflow.HttpCallTask("computeAnalytics",
				workflow.WithMethod("POST"),
				workflow.WithURI("https://api.example.com/analytics"),
				workflow.WithBody(map[string]any{
					"data": "${.data}",
					"type": "analytics",
				}),
			).Export("${.analytics}"),

			workflow.SetTask("storeAnalytics",
				workflow.SetVar("analyticsComplete", "true"),
			),
		),

		// Branch 2: Process validation
		workflow.WithBranch("validation",
			workflow.HttpCallTask("validateData",
				workflow.WithMethod("POST"),
				workflow.WithURI("https://api.example.com/validate"),
				workflow.WithBody(map[string]any{
					"data": "${.data}",
				}),
			).Export("${.validationResult}"),

			workflow.SetTask("storeValidation",
				workflow.SetVar("validationComplete", "true"),
			),
		),

		// Branch 3: Process transformation
		workflow.WithBranch("transformation",
			workflow.HttpCallTask("transformData",
				workflow.WithMethod("POST"),
				workflow.WithURI("https://api.example.com/transform"),
				workflow.WithBody(map[string]any{
					"data":   "${.data}",
					"format": "json",
				}),
			).Export("${.transformed}"),

			workflow.SetTask("storeTransformed",
				workflow.SetVar("transformationComplete", "true"),
			),
		),

		// Branch 4: Process notification
		workflow.WithBranch("notification",
			workflow.HttpCallTask("sendNotification",
				workflow.WithMethod("POST"),
				workflow.WithURI("https://api.example.com/notify"),
				workflow.WithBody(map[string]any{
					"message": "Processing started",
					"dataId":  "${.data.id}",
				}),
			),

			workflow.SetTask("markNotified",
				workflow.SetVar("notificationSent", "true"),
			),
		),
	))

	// Task 3: Aggregate results (after all branches complete)
	wf.AddTask(workflow.SetTask("aggregateResults",
		workflow.SetVar("status", "completed"),
		workflow.SetVar("analyticsStatus", "${analyticsComplete}"),
		workflow.SetVar("validationStatus", "${validationComplete}"),
		workflow.SetVar("transformationStatus", "${transformationComplete}"),
		workflow.SetVar("notificationStatus", "${notificationSent}"),
	))

	// Task 4: Send completion notification
	wf.AddTask(workflow.HttpCallTask("sendCompletion",
		workflow.WithMethod("POST"),
		workflow.WithURI("https://api.example.com/completion"),
		workflow.WithBody(map[string]any{
			"status":  "${status}",
			"results": map[string]any{
				"analytics":      "${analytics}",
				"validation":     "${validationResult}",
				"transformation": "${transformed}",
			},
		}),
	))

	log.Printf("Created parallel workflow: %s", wf)
}
