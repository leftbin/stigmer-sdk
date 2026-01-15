package workflow_test

import (
	"testing"

	"github.com/leftbin/stigmer-sdk/go/workflow"
)

func TestSetTask(t *testing.T) {
	task := workflow.SetTask("init",
		workflow.SetVar("x", "1"),
		workflow.SetVar("y", "2"),
	)

	if task.Name != "init" {
		t.Errorf("SetTask() name = %q, want %q", task.Name, "init")
	}

	if task.Kind != workflow.TaskKindSet {
		t.Errorf("SetTask() kind = %q, want %q", task.Kind, workflow.TaskKindSet)
	}

	cfg, ok := task.Config.(*workflow.SetTaskConfig)
	if !ok {
		t.Fatal("SetTask() config type is not *SetTaskConfig")
	}

	if len(cfg.Variables) != 2 {
		t.Errorf("SetTask() variables count = %d, want 2", len(cfg.Variables))
	}

	if cfg.Variables["x"] != "1" {
		t.Errorf("SetTask() variable x = %q, want %q", cfg.Variables["x"], "1")
	}

	if cfg.Variables["y"] != "2" {
		t.Errorf("SetTask() variable y = %q, want %q", cfg.Variables["y"], "2")
	}
}

func TestHttpCallTask(t *testing.T) {
	task := workflow.HttpCallTask("fetchData",
		workflow.WithMethod("GET"),
		workflow.WithURI("https://api.example.com/data"),
		workflow.WithHeader("Authorization", "Bearer ${TOKEN}"),
		workflow.WithTimeout(60),
	)

	if task.Name != "fetchData" {
		t.Errorf("HttpCallTask() name = %q, want %q", task.Name, "fetchData")
	}

	if task.Kind != workflow.TaskKindHttpCall {
		t.Errorf("HttpCallTask() kind = %q, want %q", task.Kind, workflow.TaskKindHttpCall)
	}

	cfg, ok := task.Config.(*workflow.HttpCallTaskConfig)
	if !ok {
		t.Fatal("HttpCallTask() config type is not *HttpCallTaskConfig")
	}

	if cfg.Method != "GET" {
		t.Errorf("HttpCallTask() method = %q, want %q", cfg.Method, "GET")
	}

	if cfg.URI != "https://api.example.com/data" {
		t.Errorf("HttpCallTask() uri = %q, want %q", cfg.URI, "https://api.example.com/data")
	}

	if cfg.Headers["Authorization"] != "Bearer ${TOKEN}" {
		t.Errorf("HttpCallTask() header Authorization = %q, want %q", cfg.Headers["Authorization"], "Bearer ${TOKEN}")
	}

	if cfg.TimeoutSeconds != 60 {
		t.Errorf("HttpCallTask() timeout = %d, want 60", cfg.TimeoutSeconds)
	}
}

func TestTask_Export(t *testing.T) {
	task := workflow.SetTask("init", workflow.SetVar("x", "1"))
	task.Export("${.}")

	if task.ExportAs != "${.}" {
		t.Errorf("Export() set exportAs = %q, want %q", task.ExportAs, "${.}")
	}
}

func TestTask_Then(t *testing.T) {
	task := workflow.SetTask("init", workflow.SetVar("x", "1"))
	task.Then("nextTask")

	if task.ThenTask != "nextTask" {
		t.Errorf("Then() set thenTask = %q, want %q", task.ThenTask, "nextTask")
	}
}

func TestTask_End(t *testing.T) {
	task := workflow.SetTask("init", workflow.SetVar("x", "1"))
	task.End()

	if task.ThenTask != "end" {
		t.Errorf("End() set thenTask = %q, want %q", task.ThenTask, "end")
	}
}

func TestTask_FluentAPI(t *testing.T) {
	// Test fluent API chaining
	task := workflow.HttpCallTask("fetch",
		workflow.WithMethod("GET"),
		workflow.WithURI("https://api.example.com/data"),
	).Export("${.}").Then("processData")

	if task.ExportAs != "${.}" {
		t.Errorf("Fluent API: exportAs = %q, want %q", task.ExportAs, "${.}")
	}

	if task.ThenTask != "processData" {
		t.Errorf("Fluent API: thenTask = %q, want %q", task.ThenTask, "processData")
	}
}

func TestSwitchTask(t *testing.T) {
	task := workflow.SwitchTask("checkStatus",
		workflow.WithCase("${.status == 200}", "success"),
		workflow.WithCase("${.status == 404}", "notFound"),
		workflow.WithDefault("error"),
	)

	cfg, ok := task.Config.(*workflow.SwitchTaskConfig)
	if !ok {
		t.Fatal("SwitchTask() config type is not *SwitchTaskConfig")
	}

	if len(cfg.Cases) != 2 {
		t.Errorf("SwitchTask() cases count = %d, want 2", len(cfg.Cases))
	}

	if cfg.DefaultTask != "error" {
		t.Errorf("SwitchTask() default = %q, want %q", cfg.DefaultTask, "error")
	}
}

func TestForTask(t *testing.T) {
	task := workflow.ForTask("processItems",
		workflow.WithIn("${.items}"),
		workflow.WithDo(
			workflow.SetTask("process", workflow.SetVar("item", "${.}")),
		),
	)

	cfg, ok := task.Config.(*workflow.ForTaskConfig)
	if !ok {
		t.Fatal("ForTask() config type is not *ForTaskConfig")
	}

	if cfg.In != "${.items}" {
		t.Errorf("ForTask() in = %q, want %q", cfg.In, "${.items}")
	}

	if len(cfg.Do) != 1 {
		t.Errorf("ForTask() do count = %d, want 1", len(cfg.Do))
	}
}

func TestForkTask(t *testing.T) {
	task := workflow.ForkTask("parallel",
		workflow.WithBranch("branch1",
			workflow.SetTask("task1", workflow.SetVar("x", "1")),
		),
		workflow.WithBranch("branch2",
			workflow.SetTask("task2", workflow.SetVar("y", "2")),
		),
	)

	cfg, ok := task.Config.(*workflow.ForkTaskConfig)
	if !ok {
		t.Fatal("ForkTask() config type is not *ForkTaskConfig")
	}

	if len(cfg.Branches) != 2 {
		t.Errorf("ForkTask() branches count = %d, want 2", len(cfg.Branches))
	}

	if cfg.Branches[0].Name != "branch1" {
		t.Errorf("ForkTask() branch[0] name = %q, want %q", cfg.Branches[0].Name, "branch1")
	}
}

func TestTryTask(t *testing.T) {
	task := workflow.TryTask("errorHandling",
		workflow.WithTry(
			workflow.HttpCallTask("risky", workflow.WithMethod("GET"), workflow.WithURI("${.url}")),
		),
		workflow.WithCatch([]string{"NetworkError"}, "err",
			workflow.SetTask("logError", workflow.SetVar("error", "${err}")),
		),
	)

	cfg, ok := task.Config.(*workflow.TryTaskConfig)
	if !ok {
		t.Fatal("TryTask() config type is not *TryTaskConfig")
	}

	if len(cfg.Tasks) != 1 {
		t.Errorf("TryTask() tasks count = %d, want 1", len(cfg.Tasks))
	}

	if len(cfg.Catch) != 1 {
		t.Errorf("TryTask() catch count = %d, want 1", len(cfg.Catch))
	}

	if cfg.Catch[0].As != "err" {
		t.Errorf("TryTask() catch[0] as = %q, want %q", cfg.Catch[0].As, "err")
	}
}

func TestGrpcCallTask(t *testing.T) {
	task := workflow.GrpcCallTask("callService",
		workflow.WithService("UserService"),
		workflow.WithGrpcMethod("GetUser"),
		workflow.WithGrpcBody(map[string]any{"userId": "${.userId}"}),
	)

	cfg, ok := task.Config.(*workflow.GrpcCallTaskConfig)
	if !ok {
		t.Fatal("GrpcCallTask() config type is not *GrpcCallTaskConfig")
	}

	if cfg.Service != "UserService" {
		t.Errorf("GrpcCallTask() service = %q, want %q", cfg.Service, "UserService")
	}

	if cfg.Method != "GetUser" {
		t.Errorf("GrpcCallTask() method = %q, want %q", cfg.Method, "GetUser")
	}
}

func TestListenTask(t *testing.T) {
	task := workflow.ListenTask("waitForApproval",
		workflow.WithEvent("approval.granted"),
	)

	cfg, ok := task.Config.(*workflow.ListenTaskConfig)
	if !ok {
		t.Fatal("ListenTask() config type is not *ListenTaskConfig")
	}

	if cfg.Event != "approval.granted" {
		t.Errorf("ListenTask() event = %q, want %q", cfg.Event, "approval.granted")
	}
}

func TestWaitTask(t *testing.T) {
	task := workflow.WaitTask("delay",
		workflow.WithDuration("5s"),
	)

	cfg, ok := task.Config.(*workflow.WaitTaskConfig)
	if !ok {
		t.Fatal("WaitTask() config type is not *WaitTaskConfig")
	}

	if cfg.Duration != "5s" {
		t.Errorf("WaitTask() duration = %q, want %q", cfg.Duration, "5s")
	}
}

func TestCallActivityTask(t *testing.T) {
	task := workflow.CallActivityTask("processData",
		workflow.WithActivity("DataProcessor"),
		workflow.WithActivityInput(map[string]any{"data": "${.data}"}),
	)

	cfg, ok := task.Config.(*workflow.CallActivityTaskConfig)
	if !ok {
		t.Fatal("CallActivityTask() config type is not *CallActivityTaskConfig")
	}

	if cfg.Activity != "DataProcessor" {
		t.Errorf("CallActivityTask() activity = %q, want %q", cfg.Activity, "DataProcessor")
	}
}

func TestRaiseTask(t *testing.T) {
	task := workflow.RaiseTask("throwError",
		workflow.WithError("ValidationError"),
		workflow.WithErrorMessage("Invalid input"),
	)

	cfg, ok := task.Config.(*workflow.RaiseTaskConfig)
	if !ok {
		t.Fatal("RaiseTask() config type is not *RaiseTaskConfig")
	}

	if cfg.Error != "ValidationError" {
		t.Errorf("RaiseTask() error = %q, want %q", cfg.Error, "ValidationError")
	}

	if cfg.Message != "Invalid input" {
		t.Errorf("RaiseTask() message = %q, want %q", cfg.Message, "Invalid input")
	}
}

func TestRunTask(t *testing.T) {
	task := workflow.RunTask("executeSubWorkflow",
		workflow.WithWorkflow("data-processor"),
		workflow.WithWorkflowInput(map[string]any{"data": "${.data}"}),
	)

	cfg, ok := task.Config.(*workflow.RunTaskConfig)
	if !ok {
		t.Fatal("RunTask() config type is not *RunTaskConfig")
	}

	if cfg.WorkflowName != "data-processor" {
		t.Errorf("RunTask() workflow = %q, want %q", cfg.WorkflowName, "data-processor")
	}
}

// ============================================================================
// Tests for High-Level Helpers (UX Improvements)
// ============================================================================

func TestTask_ExportAll(t *testing.T) {
	task := workflow.HttpCallTask("fetch", workflow.WithMethod("GET"), workflow.WithURI("https://api.example.com"))
	task.ExportAll()

	if task.ExportAs != "${.}" {
		t.Errorf("ExportAll() set exportAs = %q, want %q", task.ExportAs, "${.}")
	}
}

func TestTask_ExportField(t *testing.T) {
	task := workflow.HttpCallTask("fetch", workflow.WithMethod("GET"), workflow.WithURI("https://api.example.com"))
	task.ExportField("count")

	if task.ExportAs != "${.count}" {
		t.Errorf("ExportField() set exportAs = %q, want %q", task.ExportAs, "${.count}")
	}
}

func TestTask_ExportFields(t *testing.T) {
	task := workflow.HttpCallTask("fetch", workflow.WithMethod("GET"), workflow.WithURI("https://api.example.com"))
	task.ExportFields("count", "status", "data")

	// For now, ExportFields exports everything (${.})
	// In the future, it could support selective field export
	if task.ExportAs != "${.}" {
		t.Errorf("ExportFields() set exportAs = %q, want %q", task.ExportAs, "${.}")
	}
}

func TestSetInt(t *testing.T) {
	task := workflow.SetTask("init",
		workflow.SetInt("count", 42),
		workflow.SetInt("retries", 0),
	)

	cfg, ok := task.Config.(*workflow.SetTaskConfig)
	if !ok {
		t.Fatal("SetInt() config type is not *SetTaskConfig")
	}

	if cfg.Variables["count"] != "42" {
		t.Errorf("SetInt() count = %q, want %q", cfg.Variables["count"], "42")
	}

	if cfg.Variables["retries"] != "0" {
		t.Errorf("SetInt() retries = %q, want %q", cfg.Variables["retries"], "0")
	}
}

func TestSetString(t *testing.T) {
	task := workflow.SetTask("init",
		workflow.SetString("status", "pending"),
		workflow.SetString("message", "Processing..."),
	)

	cfg, ok := task.Config.(*workflow.SetTaskConfig)
	if !ok {
		t.Fatal("SetString() config type is not *SetTaskConfig")
	}

	if cfg.Variables["status"] != "pending" {
		t.Errorf("SetString() status = %q, want %q", cfg.Variables["status"], "pending")
	}

	if cfg.Variables["message"] != "Processing..." {
		t.Errorf("SetString() message = %q, want %q", cfg.Variables["message"], "Processing...")
	}
}

func TestSetBool(t *testing.T) {
	task := workflow.SetTask("init",
		workflow.SetBool("enabled", true),
		workflow.SetBool("debug", false),
	)

	cfg, ok := task.Config.(*workflow.SetTaskConfig)
	if !ok {
		t.Fatal("SetBool() config type is not *SetTaskConfig")
	}

	if cfg.Variables["enabled"] != "true" {
		t.Errorf("SetBool() enabled = %q, want %q", cfg.Variables["enabled"], "true")
	}

	if cfg.Variables["debug"] != "false" {
		t.Errorf("SetBool() debug = %q, want %q", cfg.Variables["debug"], "false")
	}
}

func TestSetFloat(t *testing.T) {
	task := workflow.SetTask("init",
		workflow.SetFloat("price", 99.99),
		workflow.SetFloat("tax", 0.08),
	)

	cfg, ok := task.Config.(*workflow.SetTaskConfig)
	if !ok {
		t.Fatal("SetFloat() config type is not *SetTaskConfig")
	}

	if cfg.Variables["price"] != "99.990000" {
		t.Errorf("SetFloat() price = %q, want %q", cfg.Variables["price"], "99.990000")
	}

	if cfg.Variables["tax"] != "0.080000" {
		t.Errorf("SetFloat() tax = %q, want %q", cfg.Variables["tax"], "0.080000")
	}
}

func TestVarRef(t *testing.T) {
	result := workflow.VarRef("apiURL")
	expected := "${apiURL}"

	if result != expected {
		t.Errorf("VarRef() = %q, want %q", result, expected)
	}
}

func TestFieldRef(t *testing.T) {
	result := workflow.FieldRef("count")
	expected := "${.count}"

	if result != expected {
		t.Errorf("FieldRef() = %q, want %q", result, expected)
	}

	// Test nested field path
	result = workflow.FieldRef("response.data.count")
	expected = "${.response.data.count}"

	if result != expected {
		t.Errorf("FieldRef() with nested path = %q, want %q", result, expected)
	}
}

func TestInterpolate(t *testing.T) {
	// Test basic interpolation
	result := workflow.Interpolate(workflow.VarRef("apiURL"), "/data")
	expected := "${apiURL}/data"

	if result != expected {
		t.Errorf("Interpolate() = %q, want %q", result, expected)
	}

	// Test multiple parts
	result = workflow.Interpolate("Bearer ", workflow.VarRef("API_TOKEN"))
	expected = "Bearer ${API_TOKEN}"

	if result != expected {
		t.Errorf("Interpolate() with Bearer = %q, want %q", result, expected)
	}

	// Test complex interpolation
	result = workflow.Interpolate(
		"https://",
		workflow.VarRef("domain"),
		"/api/v1/users/",
		workflow.FieldRef("userId"),
	)
	expected = "https://${domain}/api/v1/users/${.userId}"

	if result != expected {
		t.Errorf("Interpolate() complex = %q, want %q", result, expected)
	}
}

func TestHighLevelHelpersIntegration(t *testing.T) {
	// Test a complete workflow using all new helpers
	task := workflow.HttpCallTask("fetchUser",
		workflow.WithMethod("GET"),
		workflow.WithURI(workflow.Interpolate(workflow.VarRef("apiURL"), "/users/", workflow.FieldRef("userId"))),
		workflow.WithHeader("Authorization", workflow.Interpolate("Bearer ", workflow.VarRef("API_TOKEN"))),
	).ExportField("name")

	cfg, ok := task.Config.(*workflow.HttpCallTaskConfig)
	if !ok {
		t.Fatal("Integration test: config type is not *HttpCallTaskConfig")
	}

	expectedURI := "${apiURL}/users/${.userId}"
	if cfg.URI != expectedURI {
		t.Errorf("Integration test: URI = %q, want %q", cfg.URI, expectedURI)
	}

	expectedAuth := "Bearer ${API_TOKEN}"
	if cfg.Headers["Authorization"] != expectedAuth {
		t.Errorf("Integration test: Authorization = %q, want %q", cfg.Headers["Authorization"], expectedAuth)
	}

	expectedExport := "${.name}"
	if task.ExportAs != expectedExport {
		t.Errorf("Integration test: exportAs = %q, want %q", task.ExportAs, expectedExport)
	}
}

func TestTask_ThenRef(t *testing.T) {
	// Test type-safe task references
	task1 := workflow.SetTask("init", workflow.SetInt("x", 1))
	task2 := workflow.SetTask("process", workflow.SetInt("y", 2))
	
	task1.ThenRef(task2)

	if task1.ThenTask != "process" {
		t.Errorf("ThenRef() set thenTask = %q, want %q", task1.ThenTask, "process")
	}
}

func TestTask_EndFlow(t *testing.T) {
	// Test End() uses the EndFlow constant
	task := workflow.SetTask("final", workflow.SetString("status", "done"))
	task.End()

	if task.ThenTask != workflow.EndFlow {
		t.Errorf("End() set thenTask = %q, want %q", task.ThenTask, workflow.EndFlow)
	}

	// Verify EndFlow constant value
	if workflow.EndFlow != "end" {
		t.Errorf("EndFlow constant = %q, want %q", workflow.EndFlow, "end")
	}
}
