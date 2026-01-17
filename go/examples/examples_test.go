package examples_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	agentv1 "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/agentic/agent/v1"
	workflowv1 "buf.build/gen/go/leftbin/stigmer/protocolbuffers/go/ai/stigmer/agentic/workflow/v1"
)

// TestExample01_BasicAgent tests the basic agent example
func TestExample01_BasicAgent(t *testing.T) {
	runExampleTest(t, "01_basic_agent.go", func(t *testing.T, outputDir string) {
		// Verify agent-manifest.pb was created
		manifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		assertFileExists(t, manifestPath)

		// Verify manifest content
		var manifest agentv1.AgentManifest
		readProtoManifest(t, manifestPath, &manifest)

		// Should have 2 agents (basicAgent and fullAgent)
		if len(manifest.Agents) != 2 {
			t.Errorf("Expected 2 agents, got %d", len(manifest.Agents))
		}

		// Verify first agent
		if len(manifest.Agents) > 0 {
			agent := manifest.Agents[0]
			if agent.Name != "code-reviewer" {
				t.Errorf("First agent name = %v, want code-reviewer", agent.Name)
			}
			if agent.Instructions == "" {
				t.Error("First agent instructions should not be empty")
			}
		}

		// Verify second agent has optional fields
		if len(manifest.Agents) > 1 {
			agent := manifest.Agents[1]
			if agent.Name != "code-reviewer-pro" {
				t.Errorf("Second agent name = %v, want code-reviewer-pro", agent.Name)
			}
			if agent.Description == "" {
				t.Error("Second agent should have description")
			}
			if agent.IconUrl == "" {
				t.Error("Second agent should have icon URL")
			}
		}
	})
}

// TestExample02_AgentWithSkills tests the agent with skills example
func TestExample02_AgentWithSkills(t *testing.T) {
	runExampleTest(t, "02_agent_with_skills.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest agentv1.AgentManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Agents) < 1 {
			t.Fatal("Expected at least 1 agent")
		}

		agent := manifest.Agents[0]
		if len(agent.Skills) == 0 {
			t.Error("Agent should have skills")
		}
	})
}

// TestExample03_AgentWithMCPServers tests the agent with MCP servers example
func TestExample03_AgentWithMCPServers(t *testing.T) {
	runExampleTest(t, "03_agent_with_mcp_servers.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest agentv1.AgentManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Agents) < 1 {
			t.Fatal("Expected at least 1 agent")
		}

		agent := manifest.Agents[0]
		if len(agent.McpServers) == 0 {
			t.Error("Agent should have MCP servers")
		}
	})
}

// TestExample04_AgentWithSubagents tests the agent with subagents example
func TestExample04_AgentWithSubagents(t *testing.T) {
	runExampleTest(t, "04_agent_with_subagents.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest agentv1.AgentManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Agents) < 1 {
			t.Fatal("Expected at least 1 agent")
		}

		agent := manifest.Agents[0]
		if len(agent.SubAgents) == 0 {
			t.Error("Agent should have sub-agents")
		}
	})
}

// TestExample05_AgentWithEnvironmentVariables tests the agent with environment variables example
func TestExample05_AgentWithEnvironmentVariables(t *testing.T) {
	runExampleTest(t, "05_agent_with_environment_variables.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest agentv1.AgentManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Agents) < 1 {
			t.Fatal("Expected at least 1 agent")
		}

		agent := manifest.Agents[0]
		if len(agent.EnvironmentVariables) == 0 {
			t.Error("Agent should have environment variables")
		}
	})
}

// TestExample06_AgentWithInstructionsFromFiles tests the agent with instructions from files example
func TestExample06_AgentWithInstructionsFromFiles(t *testing.T) {
	runExampleTest(t, "06_agent_with_instructions_from_files.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest agentv1.AgentManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Agents) < 1 {
			t.Fatal("Expected at least 1 agent")
		}

		// Verify agent loaded instructions from files
		agent := manifest.Agents[0]
		if agent.Instructions == "" {
			t.Error("Agent should have instructions loaded from file")
		}
	})
}

// TestExample07_BasicWorkflow tests the basic workflow example
func TestExample07_BasicWorkflow(t *testing.T) {
	runExampleTest(t, "07_basic_workflow.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest workflowv1.WorkflowManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(manifest.Workflows))
		}

		workflow := manifest.Workflows[0]
		if workflow.Spec == nil || workflow.Spec.Document == nil {
			t.Fatal("Workflow spec or document is nil")
		}

		if workflow.Spec.Document.Namespace != "data-processing" {
			t.Errorf("Workflow namespace = %v, want data-processing", workflow.Spec.Document.Namespace)
		}
		if workflow.Spec.Document.Name != "basic-data-fetch" {
			t.Errorf("Workflow name = %v, want basic-data-fetch", workflow.Spec.Document.Name)
		}

		// Verify workflow has tasks
		if len(workflow.Spec.Tasks) == 0 {
			t.Error("Workflow should have tasks")
		}

		// Should have 3 tasks: context init + fetchData + processResponse
		// Note: Context initialization is automatically injected as first task
		if len(workflow.Spec.Tasks) != 3 {
			t.Errorf("Expected 3 tasks (including context init), got %d", len(workflow.Spec.Tasks))
		}
	})
}

// TestExample08_WorkflowWithConditionals tests the workflow with conditionals example
// TODO: This test is currently skipped because Switch/Case workflow features are not yet implemented.
// Required implementations:
//   - wf.Switch() method
//   - workflow.SwitchOn() option
//   - workflow.Case() option
//   - workflow.Equals() condition builder
//   - workflow.DefaultCase() option
//   - task.DependsOn() method
func TestExample08_WorkflowWithConditionals(t *testing.T) {
	t.Skip("TODO: Switch/Case workflow features not yet implemented (post-MVP)")

	runExampleTest(t, "08_workflow_with_conditionals.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest workflowv1.WorkflowManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(manifest.Workflows))
		}

		workflow := manifest.Workflows[0]
		if workflow.Spec == nil || workflow.Spec.Document == nil {
			t.Fatal("Workflow spec or document is nil")
		}

		if workflow.Spec.Document.Name != "conditional-deployment" {
			t.Errorf("Workflow name = %v, want conditional-deployment", workflow.Spec.Document.Name)
		}

		// Verify workflow has tasks (conditionals use SWITCH task kind)
		if len(workflow.Spec.Tasks) == 0 {
			t.Error("Workflow should have tasks")
		}
	})
}

// TestExample09_WorkflowWithLoops tests the workflow with loops example
// TODO: This test is currently skipped because ForEach/Loop workflow features are not yet implemented.
// Required implementations:
//   - wf.ForEach() method
//   - workflow.IterateOver() option
//   - workflow.WithLoopBody() option
//   - workflow.LoopVar type
//   - workflow.Body() helper (alias for WithBody)
func TestExample09_WorkflowWithLoops(t *testing.T) {
	t.Skip("TODO: ForEach/Loop workflow features not yet implemented (post-MVP)")

	runExampleTest(t, "09_workflow_with_loops.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest workflowv1.WorkflowManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(manifest.Workflows))
		}

		workflow := manifest.Workflows[0]
		if workflow.Spec == nil || workflow.Spec.Document == nil {
			t.Fatal("Workflow spec or document is nil")
		}

		if workflow.Spec.Document.Name != "batch-processor" {
			t.Errorf("Workflow name = %v, want batch-processor", workflow.Spec.Document.Name)
		}

		// Verify workflow has tasks (loops use FOR task kind)
		if len(workflow.Spec.Tasks) == 0 {
			t.Error("Workflow should have tasks")
		}
	})
}

// TestExample10_WorkflowWithErrorHandling tests the workflow with error handling example
// TODO: This test is currently skipped because Try/Catch/Finally workflow features are not yet implemented.
// Required implementations:
//   - wf.Try() method
//   - workflow.TryBlock() option
//   - workflow.CatchBlock() option
//   - workflow.FinallyBlock() option
//   - workflow.ErrorRef type
func TestExample10_WorkflowWithErrorHandling(t *testing.T) {
	t.Skip("TODO: Try/Catch/Finally workflow features not yet implemented (post-MVP)")

	runExampleTest(t, "10_workflow_with_error_handling.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest workflowv1.WorkflowManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(manifest.Workflows))
		}

		workflow := manifest.Workflows[0]
		if workflow.Spec == nil || workflow.Spec.Document == nil {
			t.Fatal("Workflow spec or document is nil")
		}

		if workflow.Spec.Document.Name != "resilient-api-call" {
			t.Errorf("Workflow name = %v, want resilient-api-call", workflow.Spec.Document.Name)
		}

		// Verify workflow has tasks (error handling uses TRY task kind)
		if len(workflow.Spec.Tasks) == 0 {
			t.Error("Workflow should have tasks")
		}
	})
}

// TestExample11_WorkflowWithParallelExecution tests the workflow with parallel execution example
// TODO: This test is currently skipped because Fork/Join parallel execution features are not yet implemented.
// Required implementations:
//   - wf.Fork() method
//   - workflow.ParallelBranches() option
//   - workflow.Branch() builder
//   - workflow.WaitForAll() option
//   - task.Branch() method to access branch results
func TestExample11_WorkflowWithParallelExecution(t *testing.T) {
	t.Skip("TODO: Fork/Join parallel execution features not yet implemented (post-MVP)")

	runExampleTest(t, "11_workflow_with_parallel_execution.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest workflowv1.WorkflowManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(manifest.Workflows))
		}

		workflow := manifest.Workflows[0]
		if workflow.Spec == nil || workflow.Spec.Document == nil {
			t.Fatal("Workflow spec or document is nil")
		}

		if workflow.Spec.Document.Name != "parallel-data-fetch" {
			t.Errorf("Workflow name = %v, want parallel-data-fetch", workflow.Spec.Document.Name)
		}

		// Verify workflow has parallel execution constructs
		if len(workflow.Spec.Tasks) == 0 {
			t.Error("Workflow should have tasks")
		}
	})
}

// TestExample12_AgentWithTypedContext tests the agent with typed context example
func TestExample12_AgentWithTypedContext(t *testing.T) {
	runExampleTest(t, "12_agent_with_typed_context.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest agentv1.AgentManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Agents) < 1 {
			t.Fatal("Expected at least 1 agent")
		}

		agent := manifest.Agents[0]
		if agent.Name != "code-reviewer" {
			t.Errorf("Agent name = %v, want code-reviewer", agent.Name)
		}

		// Verify agent has description (uses typed context variable)
		if agent.Description == "" {
			t.Error("Agent should have description from typed context")
		}

		// This example demonstrates typed context with agent
		// The key point is that typed context variables can be used for configuration
		if agent.Instructions == "" {
			t.Error("Agent should have instructions")
		}

		// Verify agent has skills (demonstrates complete agent configuration)
		if len(agent.Skills) == 0 {
			t.Error("Agent should have skills")
		}

		// Verify agent has MCP servers
		if len(agent.McpServers) == 0 {
			t.Error("Agent should have MCP servers")
		}
	})
}

// TestExample13_WorkflowAndAgentSharedContext tests the workflow and agent with shared context example
func TestExample13_WorkflowAndAgentSharedContext(t *testing.T) {
	runExampleTest(t, "13_workflow_and_agent_shared_context.go", func(t *testing.T, outputDir string) {
		// This example creates BOTH workflow and agent manifests
		workflowManifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		agentManifestPath := filepath.Join(outputDir, "agent-manifest.pb")

		// Verify both manifests were created
		assertFileExists(t, workflowManifestPath)
		assertFileExists(t, agentManifestPath)

		// Validate workflow manifest
		var workflowManifest workflowv1.WorkflowManifest
		readProtoManifest(t, workflowManifestPath, &workflowManifest)

		if len(workflowManifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(workflowManifest.Workflows))
		}

		workflow := workflowManifest.Workflows[0]
		if workflow.Spec == nil || workflow.Spec.Document == nil {
			t.Fatal("Workflow spec or document is nil")
		}

		if workflow.Spec.Document.Name != "fetch-and-analyze" {
			t.Errorf("Workflow name = %v, want fetch-and-analyze", workflow.Spec.Document.Name)
		}

		// Validate agent manifest
		var agentManifest agentv1.AgentManifest
		readProtoManifest(t, agentManifestPath, &agentManifest)

		if len(agentManifest.Agents) != 1 {
			t.Fatalf("Expected 1 agent, got %d", len(agentManifest.Agents))
		}

		agent := agentManifest.Agents[0]
		if agent.Name != "data-analyzer" {
			t.Errorf("Agent name = %v, want data-analyzer", agent.Name)
		}

		// Verify workflow has tasks (demonstrating shared context usage)
		if len(workflow.Spec.Tasks) == 0 {
			t.Error("Workflow should have tasks using shared context variables")
		}

		// Verify agent has instructions (demonstrating shared context usage)
		if agent.Instructions == "" {
			t.Error("Agent should have instructions")
		}

		// Both workflow and agent should be configured with shared context
		// This example demonstrates that both can use the same context for configuration
		// The key point is that both manifests are created successfully from the same context
	})
}

// TestExample14_AutoExportVerification tests the auto-export verification example
func TestExample14_AutoExportVerification(t *testing.T) {
	runExampleTest(t, "14_auto_export_verification.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest workflowv1.WorkflowManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(manifest.Workflows))
		}

		workflow := manifest.Workflows[0]
		if workflow.Spec == nil || workflow.Spec.Document == nil {
			t.Fatal("Workflow spec or document is nil")
		}

		if workflow.Spec.Document.Name != "auto-export-demo" {
			t.Errorf("Workflow name = %v, want auto-export-demo", workflow.Spec.Document.Name)
		}

		// Verify workflow has tasks
		if len(workflow.Spec.Tasks) == 0 {
			t.Error("Workflow should have tasks")
		}

		// Find fetchData task and verify it has export set
		var fetchDataTask *workflowv1.WorkflowTask
		for _, task := range workflow.Spec.Tasks {
			if task.Name == "fetchData" {
				fetchDataTask = task
				break
			}
		}

		if fetchDataTask == nil {
			t.Fatal("fetchData task not found")
		}

		// Verify auto-export is set
		if fetchDataTask.Export == nil {
			t.Error("fetchData task should have export (auto-export feature)")
		} else if fetchDataTask.Export.As != "${.}" {
			t.Errorf("fetchData export.as = %v, want ${.}", fetchDataTask.Export.As)
		}

		// Verify context init task exists and has export
		var contextInitTask *workflowv1.WorkflowTask
		for _, task := range workflow.Spec.Tasks {
			if task.Name == "__stigmer_init_context" {
				contextInitTask = task
				break
			}
		}

		if contextInitTask == nil {
			t.Fatal("__stigmer_init_context task not found")
		}

		// Verify context init task has export (Task 1 fix)
		if contextInitTask.Export == nil {
			t.Error("__stigmer_init_context should have export")
		} else if contextInitTask.Export.As != "${.}" {
			t.Errorf("__stigmer_init_context export.as = %v, want ${.}", contextInitTask.Export.As)
		}

		t.Log("✅ Auto-export verification: Both Task 1 and Task 2 fixes confirmed!")
	})
}

// TestExample15_AutoExportBeforeAfter tests the auto-export before/after comparison example
func TestExample15_AutoExportBeforeAfter(t *testing.T) {
	runExampleTest(t, "15_auto_export_before_after.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest workflowv1.WorkflowManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(manifest.Workflows))
		}

		workflow := manifest.Workflows[0]
		if workflow.Spec == nil || workflow.Spec.Document == nil {
			t.Fatal("Workflow spec or document is nil")
		}

		if workflow.Spec.Document.Name != "auto-export-demo" {
			t.Errorf("Workflow name = %v, want auto-export-demo", workflow.Spec.Document.Name)
		}

		// Verify workflow has tasks
		if len(workflow.Spec.Tasks) == 0 {
			t.Error("Workflow should have tasks")
		}

		// Find fetchData task and verify auto-export
		var fetchDataTask *workflowv1.WorkflowTask
		for _, task := range workflow.Spec.Tasks {
			if task.Name == "fetchData" {
				fetchDataTask = task
				break
			}
		}

		if fetchDataTask == nil {
			t.Fatal("fetchData task not found")
		}

		// Verify auto-export is set (demonstrates Task 2 fix)
		if fetchDataTask.Export == nil {
			t.Error("fetchData task should have auto-export")
		} else if fetchDataTask.Export.As != "${.}" {
			t.Errorf("fetchData export.as = %v, want ${.}", fetchDataTask.Export.As)
		}

		// Find processData task and verify implicit dependencies
		var processDataTask *workflowv1.WorkflowTask
		for _, task := range workflow.Spec.Tasks {
			if task.Name == "processData" {
				processDataTask = task
				break
			}
		}

		if processDataTask == nil {
			t.Fatal("processData task not found")
		}

		// Verify processData has variables that reference fetchData fields
		if processDataTask.TaskConfig == nil {
			t.Fatal("processData task config is nil")
		}

		varsField, ok := processDataTask.TaskConfig.Fields["variables"]
		if !ok {
			t.Fatal("processData should have 'variables' field")
		}

		varsStruct := varsField.GetStructValue()
		if varsStruct == nil {
			t.Fatal("Variables should be a struct")
		}

		// Verify field references are correct
		expectedVars := []string{"postTitle", "postBody", "postUserId", "organization"}
		for _, varName := range expectedVars {
			if _, ok := varsStruct.Fields[varName]; !ok {
				t.Errorf("Expected variable %s not found in processData task", varName)
			}
		}

		// Verify title reference contains correct expression
		if titleRef := varsStruct.Fields["postTitle"].GetStringValue(); titleRef != "${ $context.fetchData.title }" {
			t.Errorf("postTitle reference = %v, want ${ $context.fetchData.title }", titleRef)
		}

		t.Log("✅ Auto-export before/after: UX improvement verified!")
	})
}

// TestContextVariables tests the context variables example with automatic SET task injection
func TestContextVariables(t *testing.T) {
	runExampleTest(t, "context-variables/main.go", func(t *testing.T, outputDir string) {
		manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
		assertFileExists(t, manifestPath)

		var manifest workflowv1.WorkflowManifest
		readProtoManifest(t, manifestPath, &manifest)

		if len(manifest.Workflows) != 1 {
			t.Fatalf("Expected 1 workflow, got %d", len(manifest.Workflows))
		}

		workflow := manifest.Workflows[0]
		if workflow.Spec == nil {
			t.Fatal("Workflow spec is nil")
		}

		// Should have tasks: context init + user tasks
		if len(workflow.Spec.Tasks) < 2 {
			t.Fatalf("Expected at least 2 tasks (init + user tasks), got %d", len(workflow.Spec.Tasks))
		}

		// First task should be __stigmer_init_context
		initTask := workflow.Spec.Tasks[0]
		if initTask.Name != "__stigmer_init_context" {
			t.Errorf("First task name = %v, want __stigmer_init_context", initTask.Name)
		}

		// Should be a SET task
		if initTask.Kind.String() != "WORKFLOW_TASK_KIND_SET" {
			t.Errorf("Init task kind = %v, want WORKFLOW_TASK_KIND_SET", initTask.Kind)
		}

		// Verify task config has variables
		if initTask.TaskConfig == nil {
			t.Fatal("Init task config is nil")
		}

		varsField, ok := initTask.TaskConfig.Fields["variables"]
		if !ok {
			t.Fatal("Init task should have 'variables' field")
		}

		varsStruct := varsField.GetStructValue()
		if varsStruct == nil {
			t.Fatal("Variables should be a struct")
		}

		// Verify all expected variables are present
		expectedVars := []string{"apiURL", "apiVersion", "retries", "timeout", "isProd", "enableDebug", "config"}
		for _, varName := range expectedVars {
			if _, ok := varsStruct.Fields[varName]; !ok {
				t.Errorf("Expected variable %s not found in context init task", varName)
			}
		}

		// Verify variable types are correctly serialized
		// String variables
		if apiURL := varsStruct.Fields["apiURL"].GetStringValue(); apiURL != "https://api.example.com" {
			t.Errorf("apiURL = %v, want https://api.example.com", apiURL)
		}

		// Integer variables (serialized as numbers)
		if retries := varsStruct.Fields["retries"].GetNumberValue(); retries != 3 {
			t.Errorf("retries = %v, want 3", retries)
		}

		// Boolean variables
		if isProd := varsStruct.Fields["isProd"].GetBoolValue(); isProd != false {
			t.Errorf("isProd = %v, want false", isProd)
		}

		// Object variables
		configStruct := varsStruct.Fields["config"].GetStructValue()
		if configStruct == nil {
			t.Fatal("config should be a struct")
		}

		// Verify nested object structure
		dbStruct := configStruct.Fields["database"].GetStructValue()
		if dbStruct == nil {
			t.Fatal("config.database should be a struct")
		}

		if dbHost := dbStruct.Fields["host"].GetStringValue(); dbHost != "localhost" {
			t.Errorf("config.database.host = %v, want localhost", dbHost)
		}

		t.Log("✅ Context variables successfully injected with correct types!")
	})
}

// Helper function to run an example and verify output
func runExampleTest(t *testing.T, exampleFile string, verify func(*testing.T, string)) {
	t.Helper()

	// Create temporary output directory
	outputDir := t.TempDir()

	// Get the path to the example file
	examplePath := filepath.Join(".", exampleFile)

	// Check if example file exists
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Fatalf("Example file not found: %s", examplePath)
	}

	// Run the example with STIGMER_OUT_DIR set
	cmd := exec.Command("go", "run", examplePath)
	cmd.Env = append(os.Environ(), "STIGMER_OUT_DIR="+outputDir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run example %s: %v\nOutput: %s", exampleFile, err, string(output))
	}

	t.Logf("Example %s output:\n%s", exampleFile, string(output))

	// Run verification function
	verify(t, outputDir)
}

// Helper function to assert a file exists
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Expected file does not exist: %s", path)
	}
}

// Helper function to read and unmarshal a protobuf manifest
func readProtoManifest(t *testing.T, path string, message proto.Message) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read manifest file %s: %v", path, err)
	}

	if err := proto.Unmarshal(data, message); err != nil {
		t.Fatalf("Failed to unmarshal manifest %s: %v", path, err)
	}
}
