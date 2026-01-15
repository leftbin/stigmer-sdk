// Package examples demonstrates workflow conditionals using SWITCH tasks.
package main

import (
	"log"

	stigmeragent "github.com/leftbin/stigmer-sdk/go"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

// This example demonstrates conditional logic using SWITCH tasks.
//
// The workflow:
// 1. Fetches data from an API
// 2. Checks the HTTP status code
// 3. Routes to different handlers based on status
//
// Key improvements shown:
// - Version is optional (defaults to "0.1.0" if not provided)
// - Type-safe task references with .ThenRef()
// - High-level helpers (ExportAll, SetString, SetBool, FieldRef, Interpolate)
func main() {
	defer stigmeragent.Complete()

	wf, err := workflow.New(
		workflow.WithNamespace("data-processing"),
		workflow.WithName("conditional-processing"),
		// Version is optional - defaults to "0.1.0" for development
		// workflow.WithVersion("1.0.0"),  // Uncomment for production
		workflow.WithDescription("Process data with conditional routing"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Task 1: Fetch data using high-level helper
	// Returns task reference for type-safe .ThenRef()
	fetchTask := wf.AddTask(workflow.HttpCallTask("fetchData",
		workflow.WithMethod("GET"),
		workflow.WithURI("https://api.example.com/data"),
	).ExportAll()) // High-level helper instead of Export("${.}")

	// Task 2: Check HTTP status and route accordingly
	// Using string-based references (simpler but less type-safe)
	checkTask := wf.AddTask(workflow.SwitchTask("checkStatus",
		workflow.WithCase("${.status == 200}", "handleSuccess"),
		workflow.WithCase("${.status == 404}", "handleNotFound"),
		workflow.WithCase("${.status >= 500}", "handleServerError"),
		workflow.WithDefault("handleUnexpectedError"),
	))

	// Connect fetch → check using type-safe task reference
	fetchTask.ThenRef(checkTask) // Type-safe! Refactoring-friendly!

	// Task 3: Handle successful response using type-safe setters
	wf.AddTask(workflow.SetTask("handleSuccess",
		workflow.SetString("result", "success"),
		workflow.SetVar("data", workflow.FieldRef("body")),
	).End()) // Explicit .End() instead of .Then("end")

	// Task 4: Handle 404 Not Found
	wf.AddTask(workflow.SetTask("handleNotFound",
		workflow.SetString("result", "not_found"),
		workflow.SetString("message", "Resource not found"),
	).Then(workflow.EndFlow)) // Using EndFlow constant

	// Task 5: Handle server errors (5xx) using type-safe boolean
	wf.AddTask(workflow.SetTask("handleServerError",
		workflow.SetString("result", "server_error"),
		workflow.SetString("message", "Server error occurred"),
		workflow.SetBool("shouldRetry", true), // Type-safe boolean instead of "true"
	).End())

	// Task 6: Handle unexpected errors using field reference
	wf.AddTask(workflow.SetTask("handleUnexpectedError",
		workflow.SetString("result", "error"),
		workflow.SetVar("message", workflow.Interpolate("Unexpected status code: ", workflow.FieldRef("status"))),
	).End())

	log.Printf("Created conditional workflow: %s", wf)
}
