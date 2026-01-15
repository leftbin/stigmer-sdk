// Package examples demonstrates error handling using TRY/CATCH tasks.
package main

import (
	"log"

	"github.com/leftbin/stigmer-sdk/go/synthesis"
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
	defer synthesis.AutoSynth()

	wf, err := workflow.New(
		workflow.WithNamespace("data-processing"),
		workflow.WithName("error-handling"),
		workflow.WithVersion("1.0.0"),
		workflow.WithDescription("Workflow with comprehensive error handling"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Task 1: Initialize
	wf.AddTask(workflow.SetTask("initialize",
		workflow.SetVar("retryCount", "0"),
		workflow.SetVar("maxRetries", "3"),
	))

	// Task 2: Try risky operation with error handling
	wf.AddTask(workflow.TryTask("attemptDataFetch",
		// Try block - risky operation
		workflow.WithTry(
			workflow.HttpCallTask("fetchData",
				workflow.WithMethod("GET"),
				workflow.WithURI("https://api.example.com/data"),
				workflow.WithTimeout(10),
			).Export("${.}"),
		),

		// Catch network errors
		workflow.WithCatch(
			[]string{"NetworkError", "TimeoutError"},
			"networkErr",
			workflow.SetTask("handleNetworkError",
				workflow.SetVar("errorType", "network"),
				workflow.SetVar("errorMessage", "${networkErr.message}"),
				workflow.SetVar("shouldRetry", "true"),
			).Then("checkRetry"),
		),

		// Catch validation errors
		workflow.WithCatch(
			[]string{"ValidationError"},
			"validationErr",
			workflow.SetTask("handleValidationError",
				workflow.SetVar("errorType", "validation"),
				workflow.SetVar("errorMessage", "${validationErr.message}"),
				workflow.SetVar("shouldRetry", "false"),
			).Then("logError"),
		),

		// Catch all other errors
		workflow.WithCatch(
			[]string{"*"},
			"err",
			workflow.SetTask("handleUnknownError",
				workflow.SetVar("errorType", "unknown"),
				workflow.SetVar("errorMessage", "${err.message}"),
				workflow.SetVar("shouldRetry", "false"),
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

	// Task 6: Log error
	wf.AddTask(workflow.HttpCallTask("logError",
		workflow.WithMethod("POST"),
		workflow.WithURI("https://api.example.com/logs"),
		workflow.WithBody(map[string]any{
			"errorType":    "${errorType}",
			"errorMessage": "${errorMessage}",
			"retryCount":   "${retryCount}",
		}),
	))

	// Task 7: Continue with graceful degradation
	wf.AddTask(workflow.SetTask("gracefulDegradation",
		workflow.SetVar("status", "partial_failure"),
		workflow.SetVar("message", "Operation failed but workflow continued"),
	))

	log.Printf("Created error handling workflow: %s", wf)
}
