# Examples Test Suite

This directory contains comprehensive test cases for all SDK examples to ensure they generate correct manifest files.

## Test Coverage

### Agent Examples (All Passing ✅)

- `TestExample01_BasicAgent` - Tests basic agent creation with required and optional fields
- `TestExample02_AgentWithSkills` - Tests agents with inline, platform, and organization skills
- `TestExample03_AgentWithMCPServers` - Tests agents with Stdio, HTTP, and Docker MCP servers
- `TestExample04_AgentWithSubagents` - Tests agents with inline and referenced sub-agents
- `TestExample05_AgentWithEnvironmentVariables` - Tests agents with secret and config environment variables
- `TestExample06_AgentWithInstructionsFromFiles` - Tests agents that load instructions from markdown files

**Status**: 6/6 passing ✅

### Workflow Examples (Known Issues ❌)

- `TestExample07_BasicWorkflow` - Tests basic workflow with HTTP calls and variable setting
- `TestExample08_WorkflowWithConditionals` - Tests workflows with SWITCH tasks for conditional logic
- `TestExample09_WorkflowWithLoops` - Tests workflows with FOR tasks for iteration
- `TestExample10_WorkflowWithErrorHandling` - Tests workflows with TRY/CATCH error handling
- `TestExample11_WorkflowWithParallelExecution` - Tests workflows with FORK tasks for parallel execution

**Status**: 0/5 passing (SDK bug)

## Known Issues

### Workflow Proto Conversion Error

All workflow examples are currently failing with:

```
proto: invalid type: map[string]string
```

**Root Cause**: The SDK's workflow synthesis code cannot convert Go `map[string]string` types to protobuf `Struct`. This affects:

1. `SetTaskConfig.Variables` (map[string]string) - Used in SET tasks
2. `HttpCallTaskConfig.Headers` (map[string]string) - Used in HTTP_CALL tasks

**Location**: `go/workflow/task.go` lines 128 and 221

**Impact**: All workflow examples fail during manifest synthesis

**Fix Required**: The `go/internal/synth/workflow_converter.go` needs to be updated to convert these maps to `google.protobuf.Struct` or appropriate protobuf types before marshaling.

## Running Tests

```bash
# Run all tests
go test -v -timeout 180s

# Run specific test
go test -v -run TestExample01

# Run only agent tests (currently passing)
go test -v -run 'TestExample0[1-6]'

# Run only workflow tests (currently failing)
go test -v -run 'TestExample0[7-9]|TestExample1[01]'
```

## Test Implementation

### Test Pattern

Each test:

1. Creates a temporary output directory
2. Sets `STIGMER_OUT_DIR` environment variable
3. Runs the example with `go run`
4. Verifies manifest file(s) are created
5. Unmarshals and validates protobuf content
6. Checks key fields (names, descriptions, tasks, etc.)

### Example Test Structure

```go
func TestExample01_BasicAgent(t *testing.T) {
	runExampleTest(t, "01_basic_agent.go", func(t *testing.T, outputDir string) {
		// Verify agent-manifest.pb was created
		manifestPath := filepath.Join(outputDir, "agent-manifest.pb")
		assertFileExists(t, manifestPath)

		// Verify manifest content
		var manifest agentv1.AgentManifest
		readProtoManifest(t, manifestPath, &manifest)

		// Validate agents
		if len(manifest.Agents) != 2 {
			t.Errorf("Expected 2 agents, got %d", len(manifest.Agents))
		}
		// ... more assertions
	})
}
```

## Fixes Applied

### Build Tags

Added `//go:build ignore` tags to all example files to prevent package conflicts:

- `07_basic_workflow.go`
- `08_workflow_with_conditionals.go`
- `09_workflow_with_loops.go`
- `10_workflow_with_error_handling.go`
- `11_workflow_with_parallel_execution.go`
- `task3-manifest-example.go`

### Example Structure

Updated workflow examples (08-11) to pass tasks during `workflow.New()` instead of using `AddTask()` later, as the SDK requires at least one task during workflow creation:

```go
// Old pattern (causes validation error)
wf, err := workflow.New(...)  // No tasks - fails validation!
wf.AddTask(task1)
wf.AddTask(task2)

// New pattern (correct)
task1 := workflow.SetTask(...)
task2 := workflow.HttpCallTask(...)
wf, err := workflow.New(..., workflow.WithTasks(task1, task2))
```

### Unused Imports

Removed unused imports from `03_agent_with_mcp_servers.go`:
- `encoding/json`
- `google.golang.org/protobuf/encoding/protojson`

## Next Steps

1. **Fix SDK Bug**: Update `workflow_converter.go` to handle `map[string]string` conversion to protobuf
2. **Verify Workflow Tests**: Once SDK is fixed, all workflow tests should pass
3. **Expand Test Coverage**: Add more detailed assertions for workflow task configurations
4. **Golden Files**: Consider adding golden file tests for exact manifest comparison

## Test Artifacts

- Manifest files are written to `t.TempDir()` which is automatically cleaned up
- Test output shows the example's stdout/stderr for debugging
- Failed tests include the full error output from the example execution
