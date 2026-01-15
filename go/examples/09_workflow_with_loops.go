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
// Key improvements shown:
// - Type-safe task references with .ThenRef() instead of string-based .Then()
// - Type-safe switch cases with WithCaseRef() and WithDefaultRef()
// - High-level condition builders (Equals, Field, Literal)
// - Define tasks first, then reference them (more refactoring-friendly)
func main() {
	defer stigmeragent.Complete()

	wf, err := workflow.New(
		workflow.WithNamespace("data-processing"),
		workflow.WithName("batch-processing"),
		workflow.WithVersion("1.0.0"),
		workflow.WithDescription("Process multiple items in a loop"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Task 1: Fetch list of items using high-level helper
	fetchTask := workflow.HttpCallTask("fetchItems",
		workflow.WithHTTPGet(), // Type-safe HTTP method
		workflow.WithURI("https://api.example.com/items"),
	).ExportField("items")

	// Task 2: Initialize counter using type-safe setters
	initTask := workflow.SetTask("initCounter",
		workflow.SetInt("processedCount", 0),
		workflow.SetInt("failedCount", 0),
	)

	// Define counter increment tasks first so we can reference them type-safely
	//
	// Task 5: Increment success counter using expression in SetVar
	// Note: SetVar with expression allows dynamic computation like "${processedCount + 1}"
	incrementSuccessTask := workflow.SetTask("incrementSuccess",
		workflow.SetVar("processedCount", "${processedCount + 1}"),
	).End()

	// Task 6: Increment failed counter using expression in SetVar
	incrementFailedTask := workflow.SetTask("incrementFailed",
		workflow.SetVar("failedCount", "${failedCount + 1}"),
	).End()

	// Task 4: Check result and route to appropriate counter using condition builders
	//
	// Using high-level condition builder instead of raw string expression:
	// - Equals(Field("result.success"), Literal("true")) is more readable than "${.result.success}"
	// - Type-safe task references prevent typos in task names
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

	// Task 7: Aggregate results using variable references
	aggregateTask := workflow.SetTask("aggregateResults",
		workflow.SetVar("totalProcessed", workflow.VarRef("processedCount")),
		workflow.SetVar("totalFailed", workflow.VarRef("failedCount")),
		workflow.SetString("status", "completed"),
	).End()

	// Connect tasks using type-safe references
	fetchTask.ThenRef(initTask)
	initTask.ThenRef(processTask)
	processTask.ThenRef(aggregateTask)

	// Add all tasks to the workflow
	wf.AddTask(fetchTask).
		AddTask(initTask).
		AddTask(processTask).
		AddTask(checkResultTask).
		AddTask(incrementSuccessTask).
		AddTask(incrementFailedTask).
		AddTask(aggregateTask)

	log.Printf("Created loop workflow: %s", wf)
}
