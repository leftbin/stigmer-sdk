//go:build ignore

// Package examples demonstrates workflow loops using FOR tasks.
package main

import (
	"log"

	stigmeragent "github.com/leftbin/stigmer-sdk/go"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

// This example demonstrates iterating over collections using FOR tasks with modern API patterns.
//
// The workflow:
// 1. Fetches a list of items from an API
// 2. Initializes counters for tracking success/failure
// 3. Iterates over each item using FOR task
//   - Processes each item via HTTP POST
//   - Checks result using SWITCH with type-safe task references
//   - Increments success or failure counter
//
// 4. Aggregates final results
//
// Modern patterns demonstrated:
// - Type-safe HTTP methods (WithHTTPGet, WithHTTPPost) instead of raw strings
// - Arithmetic expression helpers (Increment) for counter increments
// - Variable/field reference helpers (VarRef, FieldRef) instead of "${...}" strings
// - Condition builders (Equals, Field, Literal) instead of raw expression strings
// - Type-safe task references (ThenRef, WithCaseRef, WithDefaultRef) instead of string names
// - Type-safe setters (SetInt, SetString, SetVar) for different value types
// - Export helpers (ExportField) for extracting specific response fields
// - "Define first, reference later" pattern for compile-time validation
// - FOR loop iteration over collections with type-safe task chaining
func main() {
	defer stigmeragent.Complete()

	// Task 1: Fetch list of items from API
	// ExportField("items") extracts the "items" field from response and makes it available to next task
	fetchTask := workflow.HttpCallTask("fetchItems",
		workflow.WithHTTPGet(), // Type-safe HTTP method
		workflow.WithURI("https://api.example.com/items"),
	).ExportField("items") // Make items array available for the FOR loop

	// Task 2: Initialize counter using type-safe setters
	initTask := workflow.SetTask("initCounter",
		workflow.SetInt("processedCount", 0),
		workflow.SetInt("failedCount", 0),
	)

	// Define counter increment tasks first so we can reference them type-safely
	//
	// Task 5: Increment success counter using type-safe helper
	incrementSuccessTask := workflow.SetTask("incrementSuccess",
		workflow.SetVar("processedCount", workflow.Increment("processedCount")), // ✅ Type-safe increment
	).End()

	// Task 6: Increment failed counter using type-safe helper
	incrementFailedTask := workflow.SetTask("incrementFailed",
		workflow.SetVar("failedCount", workflow.Increment("failedCount")), // ✅ Type-safe increment
	).End()

	// Task 4: Check processing result and route to appropriate counter
	//
	// This SWITCH task examines the result of processItem:
	// - If result.success == "true": increment success counter
	// - Otherwise: increment failed counter
	//
	// Using type-safe condition builder instead of raw string expression
	checkResultTask := workflow.SwitchTask("checkResult",
		workflow.WithCaseRef(
			workflow.Equals(workflow.Field("result.success"), workflow.Literal("true")),
			incrementSuccessTask,
		),
		workflow.WithDefaultRef(incrementFailedTask),
	)

	// Task 3: Process each item in a loop using field references and type-safe task references
	processTask := workflow.ForTask("processEachItem",
		workflow.WithIn(workflow.FieldRef("items")),
		workflow.WithDo(
			// Process current item using field references
			workflow.HttpCallTask("processItem",
				workflow.WithHTTPPost(), // Type-safe HTTP method
				workflow.WithURI("https://api.example.com/process"),
				workflow.WithBody(map[string]any{
					"itemId":   workflow.FieldRef("id"),
					"itemData": workflow.FieldRef("data"),
				}),
			).ExportField("result").ThenRef(checkResultTask), // Type-safe task linking

			checkResultTask,
			incrementSuccessTask,
			incrementFailedTask,
		),
	)

	// Task 7: Aggregate final results after loop completes
	// VarRef reads the counter values accumulated during loop iteration
	aggregateTask := workflow.SetTask("aggregateResults",
		workflow.SetVar("totalProcessed", workflow.VarRef("processedCount")), // Copy final count
		workflow.SetVar("totalFailed", workflow.VarRef("failedCount")),       // Copy final count
		workflow.SetString("status", "completed"),
	).End()

	// Connect tasks using type-safe references
	//
	// Flow diagram:
	//   fetchItems (GET /items → returns array)
	//        ↓
	//   initCounter (processedCount=0, failedCount=0)
	//        ↓
	//   processEachItem (FOR loop over items array)
	//        ↓ (for each item)
	//   processItem (POST /process with item data)
	//        ↓
	//   checkResult (if result.success == "true"?)
	//        ↓ YES                    ↓ NO
	//   incrementSuccess         incrementFailed
	//        ↓                        ↓
	//        └────────┬───────────────┘
	//              (loop continues for next item)
	//                 ↓
	//   aggregateResults (after all items processed)
	fetchTask.ThenRef(initTask)
	initTask.ThenRef(processTask)
	processTask.ThenRef(aggregateTask)

	// Create workflow with all tasks
	wf, err := workflow.New(
		workflow.WithNamespace("data-processing"),
		workflow.WithName("batch-processor"),
		workflow.WithVersion("1.0.0"),
		workflow.WithDescription("Process multiple items in a loop"),
		workflow.WithTasks(fetchTask, initTask, processTask, checkResultTask, incrementSuccessTask, incrementFailedTask, aggregateTask),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Created loop workflow: %s", wf)
}
