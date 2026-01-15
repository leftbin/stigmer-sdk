//go:build ignore

// Package examples demonstrates workflow conditionals using SWITCH tasks.
package main

import (
	"log"

	stigmeragent "github.com/leftbin/stigmer-sdk/go"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

// This example demonstrates conditional logic using SWITCH tasks with modern API patterns.
//
// The workflow:
// 1. Fetches data from an API (HTTP_CALL task)
//   - Returns: {status: 200, body: "...", headers: {...}}
//   - Exports all fields to workflow context with .ExportAll()
// 2. Checks the HTTP status code (SWITCH task)
//   - Accesses .status from context using Field("status")
// 3. Routes to different handlers based on status (SET tasks)
//   - Each handler can access .body, .status, etc. via FieldRef()
//
// Data flow: HTTP response → context → condition checks → handlers
//
// Modern patterns demonstrated:
// - Type-safe HTTP methods (WithHTTPGet) instead of raw strings
// - Variable/field reference helpers (FieldRef, Interpolate) instead of "${...}" strings
// - Condition builders (Equals, GreaterThanOrEqual, Field, Number) instead of raw expressions
// - Type-safe task references (ThenRef, WithCaseRef, WithDefaultRef) instead of string names
// - Type-safe setters (SetString, SetBool, SetVar) for different value types
// - Export helpers (ExportAll) for capturing entire HTTP response
// - "Define first, reference later" pattern for compile-time validation
// - Explicit .End() for terminal tasks instead of .Then("end")
// - Optional version (defaults to "0.1.0" for development)
func main() {
	defer stigmeragent.Complete()

	// Task 1: Fetch data from API
	//
	// HTTP_CALL task returns an object with these fields:
	//   - status: HTTP status code (e.g., 200, 404, 500)
	//   - body: Response body
	//   - headers: Response headers
	//
	// ExportAll() makes all response fields available to subsequent tasks
	// Using JSONPlaceholder - a free fake REST API for testing and prototyping
	fetchTask := workflow.HttpCallTask("fetchData",
		workflow.WithHTTPGet(), // Type-safe HTTP method
		workflow.WithURI("https://jsonplaceholder.typicode.com/posts/1"),
	).ExportAll() // Export entire response (status, body, headers) to context

	// Define handler tasks first for type-safe references (modern pattern)
	//
	// Task 3: Handle successful 200 response
	// FieldRef("body") accesses the body field exported by fetchTask
	successTask := workflow.SetTask("handleSuccess",
		workflow.SetString("result", "success"),
		workflow.SetVar("data", workflow.FieldRef("body")), // ✅ Type-safe field access
	).End() // Terminal task - workflow ends here

	// Task 4: Handle 404 Not Found
	notFoundTask := workflow.SetTask("handleNotFound",
		workflow.SetString("result", "not_found"),
		workflow.SetString("message", "Resource not found"),
	).End()

	// Task 5: Handle server errors (5xx)
	serverErrorTask := workflow.SetTask("handleServerError",
		workflow.SetString("result", "server_error"),
		workflow.SetString("message", "Server error occurred"),
		workflow.SetBool("shouldRetry", true), // ✅ Type-safe boolean
	).End()

	// Task 6: Handle unexpected status codes (catchall)
	// Interpolate() combines static text with dynamic field values
	unexpectedErrorTask := workflow.SetTask("handleUnexpectedError",
		workflow.SetString("result", "error"),
		workflow.SetVar("message", workflow.Interpolate("Unexpected status code: ", workflow.FieldRef("status"))), // ✅ String composition
	).End()

	// Task 2: Route based on HTTP status code
	//
	// The workflow context now contains (exported by fetchTask):
	//   - .status (HTTP status code for routing)
	//   - .body (response body accessed by handlers)
	//   - .headers (response headers)
	//
	// Using type-safe condition builders and task references
	checkTask := workflow.SwitchTask("checkStatus",
		workflow.WithCaseRef(workflow.Equals(workflow.Field("status"), workflow.Number(200)), successTask),
		workflow.WithCaseRef(workflow.Equals(workflow.Field("status"), workflow.Number(404)), notFoundTask),
		workflow.WithCaseRef(workflow.GreaterThanOrEqual(workflow.Field("status"), workflow.Number(500)), serverErrorTask),
		workflow.WithDefaultRef(unexpectedErrorTask),
	)

	// Connect fetch → check using type-safe task reference
	//
	// Flow diagram - Conditional routing based on HTTP status:
	//   fetchData (GET /data → exports status, body, headers)
	//        ↓
	//   checkStatus (SWITCH on status field)
	//        ├─ status == 200  → handleSuccess
	//        ├─ status == 404  → handleNotFound
	//        ├─ status >= 500  → handleServerError
	//        └─ default        → handleUnexpectedError
	fetchTask.ThenRef(checkTask) // Type-safe! Refactoring-friendly!

	// Create workflow with all tasks
	wf, err := workflow.New(
		workflow.WithNamespace("data-processing"),
		workflow.WithName("conditional-deployment"),
		// Version is optional - defaults to "0.1.0" for development
		// workflow.WithVersion("1.0.0"),  // Uncomment for production
		workflow.WithDescription("Process data with conditional routing"),
		workflow.WithTasks(fetchTask, checkTask, successTask, notFoundTask, serverErrorTask, unexpectedErrorTask),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Created conditional workflow: %s", wf)
}
