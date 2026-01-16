# Examples Test Suite

This directory contains comprehensive test cases for all SDK examples to ensure they generate correct manifest files.

**Last Updated**: 2026-01-17  
**Status**: ✅ **100% Coverage - All Tests Passing**

## Test Coverage Summary

- **Total Tests**: 13
- **Passing**: 13/13 (100%) ✅
- **Runtime**: ~2.15 seconds
- **Coverage**: All user-facing examples

## Test Coverage by Category

### Agent Examples (7 tests) ✅

- `TestExample01_BasicAgent` - Tests basic agent creation with required and optional fields
- `TestExample02_AgentWithSkills` - Tests agents with inline, platform, and organization skills
- `TestExample03_AgentWithMCPServers` - Tests agents with Stdio, HTTP, and Docker MCP servers
- `TestExample04_AgentWithSubagents` - Tests agents with inline and referenced sub-agents
- `TestExample05_AgentWithEnvironmentVariables` - Tests agents with secret and config environment variables
- `TestExample06_AgentWithInstructionsFromFiles` - Tests agents that load instructions from markdown files
- `TestExample08_AgentWithTypedContext` - Tests agent with typed context variables (NEW API) ⭐ **NEW!**

**Status**: 7/7 passing ✅

### Workflow Examples - NEW API (2 tests) ✅

- `TestExample07_BasicWorkflow` - Tests basic workflow with HTTP calls, typed context, and implicit dependencies (NEW API)
- `TestExample09_WorkflowAndAgentSharedContext` - Tests workflow and agent sharing the same typed context ⭐ **NEW!**

**Status**: 2/2 passing ✅

### Workflow Examples - OLD API (4 tests) ✅

- `TestExample08_WorkflowWithConditionals` - Tests workflows with SWITCH tasks for conditional logic
- `TestExample09_WorkflowWithLoops` - Tests workflows with FOR tasks for iteration
- `TestExample10_WorkflowWithErrorHandling` - Tests workflows with TRY/CATCH error handling
- `TestExample11_WorkflowWithParallelExecution` - Tests workflows with FORK tasks for parallel execution

**Status**: 4/4 passing ✅  
**Note**: These examples use the old API and are marked with warning headers

## Recent Updates (2026-01-17)

### New Tests Added ⭐

**1. TestExample08_AgentWithTypedContext**
- **File**: `08_agent_with_typed_context.go`
- **Purpose**: Validates agent creation with typed context variables
- **Tests**: 
  - Agent manifest generation
  - Typed context variable usage (agentName, iconURL, org)
  - Skills and MCP servers configuration
  - Demonstrates NEW Pulumi-aligned API

**2. TestExample09_WorkflowAndAgentSharedContext**
- **File**: `09_workflow_and_agent_shared_context.go`
- **Purpose**: Validates workflow and agent sharing the same typed context
- **Tests**:
  - Both workflow and agent manifest generation
  - Shared context variables (apiURL, orgName, retryCount)
  - Demonstrates advanced NEW API pattern

### Obsolete Files Removed

- Deleted `task3-manifest-example.go` (obsolete internal reference file)

## Running Tests

```bash
# Run all tests (recommended)
go test -v

# Run all tests without cache (for debugging)
go test -v -count=1

# Run specific test
go test -v -run TestExample01_BasicAgent

# Run only agent tests
go test -v -run 'TestExample0[1-6]|TestExample08_AgentWithTypedContext'

# Run only NEW API workflow tests
go test -v -run 'TestExample07_BasicWorkflow|TestExample09_WorkflowAndAgentSharedContext'

# Run only OLD API workflow tests
go test -v -run 'TestExample0[8-9]_Workflow|TestExample1[01]'
```

**Expected Output**:
```
PASS
ok  	github.com/leftbin/stigmer-sdk/go/examples	2.151s
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

## API Migration Status

### NEW API Examples (Pulumi-Aligned) ✅

These examples use the new Pulumi-aligned API with `stigmer.Run()` and typed context:

- `07_basic_workflow.go` - Demonstrates HTTP tasks, field references, implicit dependencies
- `08_agent_with_typed_context.go` - Demonstrates agent with typed context variables
- `09_workflow_and_agent_shared_context.go` - Demonstrates shared context pattern

**Features**:
- ✅ `stigmer.Run(func(ctx) {...})` pattern
- ✅ Typed context variables
- ✅ `task.Field()` for task output references
- ✅ Implicit dependency tracking
- ✅ Clean workflow builders (`wf.HttpGet()`, etc.)

### OLD API Examples (Marked as Legacy) ⚠️

These examples use the old API and have warning headers:

- `08_workflow_with_conditionals.go` - SWITCH tasks
- `09_workflow_with_loops.go` - FOR tasks
- `10_workflow_with_error_handling.go` - TRY tasks
- `11_workflow_with_parallel_execution.go` - FORK tasks
- `07_basic_workflow_legacy.go` - Intentional reference

**Status**: All have `//go:build ignore` tags and warning headers pointing to migration guide.

## Next Steps

### Short-term (Optional - Phase 5.3)
- Migrate OLD API workflow examples (08-11) to NEW API
- Remove warning headers from migrated examples

### Long-term (Phase 7)
- Integration testing with workflow-runner
- Integration testing with agent-runner
- End-to-end execution validation
- Performance testing

### Quality Improvements (Future)
- Add golden file tests for exact manifest comparison
- Expand test assertions for workflow task configurations
- Add benchmark tests

## Test Artifacts

- Manifest files are written to `t.TempDir()` which is automatically cleaned up
- Test output shows the example's stdout/stderr for debugging
- Failed tests include the full error output from the example execution
