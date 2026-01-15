//go:build ignore

// Package examples demonstrates parallel execution using FORK tasks.
package main

import (
	"log"

	stigmeragent "github.com/leftbin/stigmer-sdk/go"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

// This example demonstrates parallel task execution using FORK tasks with modern API patterns.
//
// The workflow:
// 1. Fetches data that needs parallel processing
// 2. Forks into 4 parallel branches (analytics, validation, transformation, notification)
// 3. Each branch processes data independently and concurrently
// 4. Automatically joins when all branches complete
// 5. Aggregates results from all branches
// 6. Sends completion notification
//
// Modern patterns demonstrated:
// - Type-safe HTTP methods (WithHTTPGet, WithHTTPPost) instead of raw strings
// - Variable/field reference helpers (VarRef, FieldRef) instead of "${...}" strings
// - Type-safe setters (SetBool, SetString, SetVar) for different value types
// - Export helpers (ExportAll, ExportField) for capturing response data
// - FORK task for parallel execution with named branches
// - Automatic join semantics (next task waits for all branches to complete)
// - Nested task structures within branches
func main() {
	defer stigmeragent.Complete()

	// Task 1: Fetch data to process
	// Using JSONPlaceholder - a free fake REST API for testing and prototyping
	// ExportAll() makes entire response available to parallel branches
	fetchTask := workflow.HttpCallTask("fetchData",
		workflow.WithHTTPGet(), // Type-safe HTTP method
		workflow.WithURI("https://jsonplaceholder.typicode.com/posts/1"),
	).ExportAll() // All branches can access this data via FieldRef("title"), FieldRef("body"), etc.

	// Task 2: Fork into 4 parallel branches
	//
	// Flow diagram - Parallel execution:
	//   fetchData (GET /data → exports to all branches)
	//        ↓
	//   parallelProcessing (FORK - all branches execute concurrently)
	//        ├─ analytics branch      (POST /posts → export result)
	//        ├─ validation branch     (POST /posts → export result)
	//        ├─ transformation branch (POST /posts → export result)
	//        └─ notification branch   (POST /posts → send alert)
	//        ↓ (automatic join - waits for ALL branches to complete)
	//   aggregateResults (collect all branch results)
	//        ↓
	//   sendCompletion (POST /posts with aggregated data)
	forkTask := workflow.ForkTask("parallelProcessing",
		// Branch 1: Analytics processing (runs concurrently with other branches)
		workflow.WithBranch("analytics",
			workflow.HttpCallTask("computeAnalytics",
				workflow.WithHTTPPost(), // Type-safe HTTP method
				workflow.WithURI("https://jsonplaceholder.typicode.com/posts"),
				workflow.WithBody(map[string]any{
					"title": "Analytics Result",
					"body":  workflow.Interpolate("Analyzing post: ", workflow.FieldRef("title")),
					"type":  "analytics",
				}),
			).ExportField("id"), // Make result available to aggregateResults

			workflow.SetTask("storeAnalytics",
				workflow.SetBool("analyticsComplete", true), // ✅ Mark branch complete
			),
		),

		// Branch 2: Validation processing (independent of other branches)
		workflow.WithBranch("validation",
			workflow.HttpCallTask("validateData",
				workflow.WithHTTPPost(), // Type-safe HTTP method
				workflow.WithURI("https://jsonplaceholder.typicode.com/posts"),
				workflow.WithBody(map[string]any{
					"title": "Validation Result",
					"body":  workflow.Interpolate("Validating post: ", workflow.FieldRef("title")),
				}),
			).ExportField("id"), // Export validation result

			workflow.SetTask("storeValidation",
				workflow.SetBool("validationComplete", true), // ✅ Mark branch complete
			),
		),

		// Branch 3: Transformation processing (independent of other branches)
		workflow.WithBranch("transformation",
			workflow.HttpCallTask("transformData",
				workflow.WithHTTPPost(), // Type-safe HTTP method
				workflow.WithURI("https://jsonplaceholder.typicode.com/posts"),
				workflow.WithBody(map[string]any{
					"title":  "Transformation Result",
					"body":   workflow.Interpolate("Transformed: ", workflow.FieldRef("body")),
					"format": "json",
				}),
			).ExportField("id"), // Export transformed data ID

			workflow.SetTask("storeTransformed",
				workflow.SetBool("transformationComplete", true), // ✅ Mark branch complete
			),
		),

		// Branch 4: Notification processing (fastest branch - no heavy computation)
		workflow.WithBranch("notification",
			workflow.HttpCallTask("sendNotification",
				workflow.WithHTTPPost(), // Type-safe HTTP method
				workflow.WithURI("https://jsonplaceholder.typicode.com/posts"),
				workflow.WithBody(map[string]any{
					"title":   "Notification",
					"body":    workflow.Interpolate("Processing started for: ", workflow.FieldRef("title")),
					"message": "Processing started",
				}),
			),

			workflow.SetTask("markNotified",
				workflow.SetBool("notificationSent", true), // ✅ Mark branch complete
			),
		),
	)

	// Task 3: Aggregate results (executes after ALL parallel branches complete)
	// VarRef reads the variables set by each branch
	aggregateTask := workflow.SetTask("aggregateResults",
		workflow.SetString("status", "completed"),
		workflow.SetVar("analyticsStatus", workflow.VarRef("analyticsComplete")),           // ✅ From branch 1
		workflow.SetVar("validationStatus", workflow.VarRef("validationComplete")),         // ✅ From branch 2
		workflow.SetVar("transformationStatus", workflow.VarRef("transformationComplete")), // ✅ From branch 3
		workflow.SetVar("notificationStatus", workflow.VarRef("notificationSent")),         // ✅ From branch 4
	)

	// Task 4: Send completion notification with all results
	// VarRef accesses both status variables and exported results from branches
	completionTask := workflow.HttpCallTask("sendCompletion",
		workflow.WithHTTPPost(), // Type-safe HTTP method
		workflow.WithURI("https://jsonplaceholder.typicode.com/posts"),
		workflow.WithBody(map[string]any{
			"title":  "All Processing Complete",
			"status": workflow.VarRef("status"), // ✅ Aggregation status
			"body":   "All parallel branches completed successfully",
		}),
	)

	// Connect tasks
	fetchTask.ThenRef(forkTask)
	forkTask.ThenRef(aggregateTask)
	aggregateTask.ThenRef(completionTask)

	// Create workflow with all tasks
	wf, err := workflow.New(
		workflow.WithNamespace("data-processing"),
		workflow.WithName("parallel-data-fetch"),
		workflow.WithVersion("1.0.0"),
		workflow.WithDescription("Process data in parallel branches"),
		workflow.WithTasks(fetchTask, forkTask, aggregateTask, completionTask),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Created parallel workflow: %s", wf)
}
