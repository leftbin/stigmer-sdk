// Package examples demonstrates workflow loops using FOR tasks.
package main

import (
	"log"

	"github.com/leftbin/stigmer-sdk/go/synthesis"
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
	defer synthesis.AutoSynth()

	wf, err := workflow.New(
		workflow.WithNamespace("data-processing"),
		workflow.WithName("batch-processing"),
		workflow.WithVersion("1.0.0"),
		workflow.WithDescription("Process multiple items in a loop"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Task 1: Fetch list of items
	wf.AddTask(workflow.HttpCallTask("fetchItems",
		workflow.WithMethod("GET"),
		workflow.WithURI("https://api.example.com/items"),
	).Export("${.items}"))

	// Task 2: Initialize counter
	wf.AddTask(workflow.SetTask("initCounter",
		workflow.SetVar("processedCount", "0"),
		workflow.SetVar("failedCount", "0"),
	))

	// Task 3: Process each item in a loop
	wf.AddTask(workflow.ForTask("processEachItem",
		workflow.WithIn("${.items}"),
		workflow.WithDo(
			// Process current item
			workflow.HttpCallTask("processItem",
				workflow.WithMethod("POST"),
				workflow.WithURI("https://api.example.com/process"),
				workflow.WithBody(map[string]any{
					"itemId":   "${.id}",
					"itemData": "${.data}",
				}),
			).Export("${.result}"),

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

	// Task 4: Aggregate results
	wf.AddTask(workflow.SetTask("aggregateResults",
		workflow.SetVar("totalProcessed", "${processedCount}"),
		workflow.SetVar("totalFailed", "${failedCount}"),
		workflow.SetVar("status", "completed"),
	))

	log.Printf("Created loop workflow: %s", wf)
}
