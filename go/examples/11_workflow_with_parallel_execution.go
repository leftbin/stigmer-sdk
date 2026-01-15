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
	// ExportAll() makes entire response available to parallel branches
	fetchTask := workflow.HttpCallTask("fetchData",
		workflow.WithHTTPGet(), // Type-safe HTTP method
		workflow.WithURI("https://api.example.com/data"),
	).ExportAll() // All branches can access this data via FieldRef("data")

	// Task 2: Fork into 4 parallel branches
	//
	// Flow diagram - Parallel execution:
	//   fetchData (GET /data → exports to all branches)
	//        ↓
	//   parallelProcessing (FORK - all branches execute concurrently)
	//        ├─ analytics branch      (POST /analytics → export result)
	//        ├─ validation branch     (POST /validate → export result)
	//        ├─ transformation branch (POST /transform → export result)
	//        └─ notification branch   (POST /notify → send alert)
	//        ↓ (automatic join - waits for ALL branches to complete)
	//   aggregateResults (collect all branch results)
	//        ↓
	//   sendCompletion (POST /completion with aggregated data)
	forkTask := workflow.ForkTask("parallelProcessing",
		// Branch 1: Analytics processing (runs concurrently with other branches)
		workflow.WithBranch("analytics",
			workflow.HttpCallTask("computeAnalytics",
				workflow.WithHTTPPost(), // Type-safe HTTP method
				workflow.WithURI("https://api.example.com/analytics"),
				workflow.WithBody(map[string]any{
					"data": workflow.FieldRef("data"), // ✅ Access fetched data
					"type": "analytics",
				}),
			).ExportField("analytics"), // Make result available to aggregateResults

			workflow.SetTask("storeAnalytics",
				workflow.SetBool("analyticsComplete", true), // ✅ Mark branch complete
			),
		),

		// Branch 2: Validation processing (independent of other branches)
		workflow.WithBranch("validation",
			workflow.HttpCallTask("validateData",
				workflow.WithHTTPPost(), // Type-safe HTTP method
				workflow.WithURI("https://api.example.com/validate"),
				workflow.WithBody(map[string]any{
					"data": workflow.FieldRef("data"), // ✅ Access fetched data
				}),
			).ExportField("validationResult"), // Export validation result

			workflow.SetTask("storeValidation",
				workflow.SetBool("validationComplete", true), // ✅ Mark branch complete
			),
		),

		// Branch 3: Transformation processing (independent of other branches)
		workflow.WithBranch("transformation",
			workflow.HttpCallTask("transformData",
				workflow.WithHTTPPost(), // Type-safe HTTP method
				workflow.WithURI("https://api.example.com/transform"),
				workflow.WithBody(map[string]any{
					"data":   workflow.FieldRef("data"), // ✅ Access fetched data
					"format": "json",
				}),
			).ExportField("transformed"), // Export transformed data

			workflow.SetTask("storeTransformed",
				workflow.SetBool("transformationComplete", true), // ✅ Mark branch complete
			),
		),

		// Branch 4: Notification processing (fastest branch - no heavy computation)
		workflow.WithBranch("notification",
			workflow.HttpCallTask("sendNotification",
				workflow.WithHTTPPost(), // Type-safe HTTP method
				workflow.WithURI("https://api.example.com/notify"),
				workflow.WithBody(map[string]any{
					"message": "Processing started",
					"dataId":  workflow.FieldRef("data.id"), // ✅ Access nested field
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
		workflow.WithURI("https://api.example.com/completion"),
		workflow.WithBody(map[string]any{
			"status": workflow.VarRef("status"), // ✅ Aggregation status
			"results": map[string]any{
				"analytics":      workflow.VarRef("analytics"),        // ✅ From analytics branch
				"validation":     workflow.VarRef("validationResult"), // ✅ From validation branch
				"transformation": workflow.VarRef("transformed"),      // ✅ From transformation branch
			},
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
