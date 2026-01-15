// Package examples demonstrates workflow loops using FOR tasks.
package main

import (
	"log"

	stigmeragent "github.com/leftbin/stigmer-sdk/go"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

// This example demonstrates iterating over collections using FOR tasks.
//
// The workflow:
// 1. Fetches a list of items from an API
// 2. Iterates over each item
// 3. Processes each item individually
// 4. Aggregates results
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
	wf.AddTask(workflow.HttpCallTask("fetchItems",
		workflow.WithMethod("GET"),
		workflow.WithURI("https://api.example.com/items"),
	).ExportField("items"))

	// Task 2: Initialize counter using type-safe setters
	wf.AddTask(workflow.SetTask("initCounter",
		workflow.SetInt("processedCount", 0),
		workflow.SetInt("failedCount", 0),
	))

	// Task 3: Process each item in a loop using field references
	wf.AddTask(workflow.ForTask("processEachItem",
		workflow.WithIn(workflow.FieldRef("items")),
		workflow.WithDo(
			// Process current item using field references
			workflow.HttpCallTask("processItem",
				workflow.WithMethod("POST"),
				workflow.WithURI("https://api.example.com/process"),
				workflow.WithBody(map[string]any{
					"itemId":   workflow.FieldRef("id"),
					"itemData": workflow.FieldRef("data"),
				}),
			).ExportField("result"),

			// Update counter based on result
			workflow.SwitchTask("checkResult",
				workflow.WithCase("${.result.success}", "incrementSuccess"),
				workflow.WithDefault("incrementFailed"),
			),

			workflow.SetTask("incrementSuccess",
				workflow.SetVar("processedCount", "${processedCount + 1}"),
			).Then("end"),

			workflow.SetTask("incrementFailed",
				workflow.SetVar("failedCount", "${failedCount + 1}"),
			).End(),
		),
	))

	// Task 4: Aggregate results using variable references
	wf.AddTask(workflow.SetTask("aggregateResults",
		workflow.SetVar("totalProcessed", workflow.VarRef("processedCount")),
		workflow.SetVar("totalFailed", workflow.VarRef("failedCount")),
		workflow.SetString("status", "completed"),
	))

	log.Printf("Created loop workflow: %s", wf)
}
