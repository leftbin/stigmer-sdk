// Package examples demonstrates workflow conditionals using SWITCH tasks.
package main

import (
	"log"

	"github.com/leftbin/stigmer-sdk/go/synthesis"
	"github.com/leftbin/stigmer-sdk/go/workflow"
)

// This example demonstrates conditional logic using SWITCH tasks.
//
// The workflow:
// 1. Fetches data from an API
// 2. Checks the HTTP status code
// 3. Routes to different handlers based on status
func main() {
	defer synthesis.AutoSynth()

	wf, err := workflow.New(
		workflow.WithNamespace("data-processing"),
		workflow.WithName("conditional-processing"),
		workflow.WithVersion("1.0.0"),
		workflow.WithDescription("Process data with conditional routing"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Task 1: Fetch data
	wf.AddTask(workflow.HttpCallTask("fetchData",
		workflow.WithMethod("GET"),
		workflow.WithURI("https://api.example.com/data"),
	).Export("${.}").Then("checkStatus"))

	// Task 2: Check HTTP status and route accordingly
	wf.AddTask(workflow.SwitchTask("checkStatus",
		workflow.WithCase("${.status == 200}", "handleSuccess"),
		workflow.WithCase("${.status == 404}", "handleNotFound"),
		workflow.WithCase("${.status >= 500}", "handleServerError"),
		workflow.WithDefault("handleUnexpectedError"),
	))

	// Task 3: Handle successful response
	wf.AddTask(workflow.SetTask("handleSuccess",
		workflow.SetVar("result", "success"),
		workflow.SetVar("data", "${.body}"),
	).Then("end"))

	// Task 4: Handle 404 Not Found
	wf.AddTask(workflow.SetTask("handleNotFound",
		workflow.SetVar("result", "not_found"),
		workflow.SetVar("message", "Resource not found"),
	).Then("end"))

	// Task 5: Handle server errors (5xx)
	wf.AddTask(workflow.SetTask("handleServerError",
		workflow.SetVar("result", "server_error"),
		workflow.SetVar("message", "Server error occurred"),
		workflow.SetVar("shouldRetry", "true"),
	).Then("end"))

	// Task 6: Handle unexpected errors
	wf.AddTask(workflow.SetTask("handleUnexpectedError",
		workflow.SetVar("result", "error"),
		workflow.SetVar("message", "Unexpected status code: ${.status}"),
	).End())

	log.Printf("Created conditional workflow: %s", wf)
}
