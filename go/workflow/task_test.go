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
