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
// 1. Fetches data from an API (HTTP_CALL task)
//   - Returns: {status: 200, body: "...", headers: {...}}
//   - Exports all fields to workflow context with .ExportAll()
//
// 2. Checks the HTTP status code (SWITCH task)
//   - Accesses .status from context using Field("status")
//
// 3. Routes to different handlers based on status (SET tasks)
//   - Each handler can access .body, .status, etc. via FieldRef()
//
// Data flow: HTTP response → context → condition checks → handlers
//
// Key improvements shown:
// - Version is optional (defaults to "0.1.0" if not provided)
// - Type-safe task references with .ThenRef()
// - Type-safe switch cases with WithCaseRef() and WithDefaultRef()
// - High-level condition builders (Equals, GreaterThanOrEqual, Field, Number)
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
	//
	// HTTP_CALL task returns an object with these fields:
	//   - status: HTTP status code (e.g., 200, 404, 500)
	//   - body: Response body
	//   - headers: Response headers
	//
	// .ExportAll() exports this entire object to the workflow context,
	// making all fields available to subsequent tasks via Field("fieldName")
	fetchTask := workflow.HttpCallTask("fetchData",
		workflow.WithHTTPGet(), // Type-safe HTTP method
		workflow.WithURI("https://api.example.com/data"),
	).ExportAll() // Exports the HTTP response (status, body, headers) to context

	// Define handler tasks first so we can reference them type-safely
	// Task 3: Handle successful response using type-safe setters
	//
	// FieldRef("body") accesses the "body" field from the HTTP response
	// that was exported by fetchTask.ExportAll()
	successTask := workflow.SetTask("handleSuccess",
		workflow.SetString("result", "success"),
		workflow.SetVar("data", workflow.FieldRef("body")), // Access HTTP response body
	).End() // Explicit .End() instead of .Then("end")

	// Task 4: Handle 404 Not Found
	notFoundTask := workflow.SetTask("handleNotFound",
		workflow.SetString("result", "not_found"),
		workflow.SetString("message", "Resource not found"),
	).End()

	// Task 5: Handle server errors (5xx) using type-safe boolean
	serverErrorTask := workflow.SetTask("handleServerError",
		workflow.SetString("result", "server_error"),
		workflow.SetString("message", "Server error occurred"),
		workflow.SetBool("shouldRetry", true), // Type-safe boolean instead of "true"
	).End()

	// Task 6: Handle unexpected errors using field reference
	//
	// FieldRef("status") accesses the HTTP status code from the exported response
	// Interpolate() combines the literal string with the status value
	unexpectedErrorTask := workflow.SetTask("handleUnexpectedError",
		workflow.SetString("result", "error"),
		workflow.SetVar("message", workflow.Interpolate("Unexpected status code: ", workflow.FieldRef("status"))),
	).End()

	// Task 2: Check HTTP status and route accordingly
	//
	// Field("status") accesses the "status" field from the HTTP response that was
	// exported by fetchTask.ExportAll() above. The workflow context now contains:
	//   - .status (the HTTP status code we're checking)
	//   - .body (the response body we reference in handlers)
	//   - .headers (the response headers)
	//
	// Using high-level condition builders + type-safe task references
	checkTask := workflow.SwitchTask("checkStatus",
		workflow.WithCaseRef(workflow.Equals(workflow.Field("status"), workflow.Number(200)), successTask),
		workflow.WithCaseRef(workflow.Equals(workflow.Field("status"), workflow.Number(404)), notFoundTask),
		workflow.WithCaseRef(workflow.GreaterThanOrEqual(workflow.Field("status"), workflow.Number(500)), serverErrorTask),
		workflow.WithDefaultRef(unexpectedErrorTask),
	)

	// Connect fetch → check using type-safe task reference
	fetchTask.ThenRef(checkTask) // Type-safe! Refactoring-friendly!

	// Add all tasks to the workflow
	wf.AddTask(fetchTask).AddTask(checkTask).AddTask(successTask).AddTask(notFoundTask).AddTask(serverErrorTask).AddTask(unexpectedErrorTask)

	log.Printf("Created conditional workflow: %s", wf)
}
