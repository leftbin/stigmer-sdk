// Package examples demonstrates error handling using TRY/CATCH tasks.
package main

import (
	"log"

	stigmeragent "github.com/leftbin/stigmer-sdk/go"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

// This example demonstrates error handling using TRY tasks.
//
// The workflow:
// 1. Attempts a risky operation (HTTP call)
// 2. Catches network errors and handles them
// 3. Catches validation errors separately
// 4. Logs errors and continues execution
func main() {
	defer stigmeragent.Complete()

	wf, err := workflow.New(
		workflow.WithNamespace("data-processing"),
		workflow.WithName("error-handling"),
		workflow.WithVersion("1.0.0"),
		workflow.WithDescription("Workflow with comprehensive error handling"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Task 1: Initialize using type-safe setters
	wf.AddTask(workflow.SetTask("initialize",
		workflow.SetInt("retryCount", 0),
		workflow.SetInt("maxRetries", 3),
	))

	// Task 2: Try risky operation with error handling using high-level helpers
	wf.AddTask(workflow.TryTask("attemptDataFetch",
		// Try block - risky operation
		workflow.WithTry(
			workflow.HttpCallTask("fetchData",
				workflow.WithHTTPGet(), // Type-safe HTTP method
				workflow.WithURI("https://api.example.com/data"),
				workflow.WithTimeout(10),
			).ExportAll(),
		),

		// Catch network errors using type-safe boolean
		workflow.WithCatch(
			[]string{"NetworkError", "TimeoutError"},
			"networkErr",
			workflow.SetTask("handleNetworkError",
				workflow.SetString("errorType", "network"),
				workflow.SetVar("errorMessage", "${networkErr.message}"),
				workflow.SetBool("shouldRetry", true),
			).Then("checkRetry"),
		),

		// Catch validation errors using type-safe boolean
		workflow.WithCatch(
			[]string{"ValidationError"},
			"validationErr",
			workflow.SetTask("handleValidationError",
				workflow.SetString("errorType", "validation"),
				workflow.SetVar("errorMessage", "${validationErr.message}"),
				workflow.SetBool("shouldRetry", false),
			).Then("logError"),
		),

		// Catch all other errors using type-safe boolean
		workflow.WithCatch(
			[]string{"*"},
			"err",
			workflow.SetTask("handleUnknownError",
				workflow.SetString("errorType", "unknown"),
				workflow.SetVar("errorMessage", "${err.message}"),
				workflow.SetBool("shouldRetry", false),
			).Then("logError"),
		),
	))

	// Task 3: Check if should retry
	wf.AddTask(workflow.SwitchTask("checkRetry",
		workflow.WithCase("${shouldRetry && retryCount < maxRetries}", "retry"),
		workflow.WithDefault("logError"),
	))

	// Task 4: Retry the operation
	wf.AddTask(workflow.SetTask("retry",
		workflow.SetVar("retryCount", "${retryCount + 1}"),
	).Then("waitBeforeRetry"))

	// Task 5: Wait before retry
	wf.AddTask(workflow.WaitTask("waitBeforeRetry",
		workflow.WithDuration("5s"),
	).Then("attemptDataFetch"))

	// Task 6: Log error using variable references
	wf.AddTask(workflow.HttpCallTask("logError",
		workflow.WithHTTPPost(), // Type-safe HTTP method
		workflow.WithURI("https://api.example.com/logs"),
		workflow.WithBody(map[string]any{
			"errorType":    workflow.VarRef("errorType"),
			"errorMessage": workflow.VarRef("errorMessage"),
			"retryCount":   workflow.VarRef("retryCount"),
		}),
	))

	// Task 7: Continue with graceful degradation using type-safe setters
	wf.AddTask(workflow.SetTask("gracefulDegradation",
		workflow.SetString("status", "partial_failure"),
		workflow.SetString("message", "Operation failed but workflow continued"),
	))

	log.Printf("Created error handling workflow: %s", wf)
}
