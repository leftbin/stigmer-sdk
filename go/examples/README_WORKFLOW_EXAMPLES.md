# Workflow Examples Status

This document tracks the status of workflow examples and their API versions.

## Updated Examples (NEW Pulumi-Aligned API)

### ✅ 07_basic_workflow.go
- **Status**: Fully updated to new API
- **Features**: HTTP GET, task field references, implicit dependencies
- **API**: stigmer.Run(), wf.HttpGet(), task.Field()
- **Recommended**: ⭐ **START HERE** for learning the new API

### ✅ 08_agent_with_typed_context.go
- **Status**: Fully updated to new API
- **Features**: Agent with typed context variables
- **API**: stigmer.Run(), agent.NewWithContext()

### ✅ 09_workflow_and_agent_shared_context.go
- **Status**: Fully updated (shared context pattern)
- **Features**: Workflow and agent sharing configuration
- **API**: stigmer.Run(), ctx shared between workflow and agent

## Legacy Examples (OLD API - Need Migration)

The following examples use the OLD API patterns and need to be updated:

### ⚠️ 08_workflow_with_conditionals.go
- **Status**: Uses OLD API
- **Features**: SWITCH tasks, conditional logic
- **Old patterns**: 
  - `defer stigmer.Complete()` (should use `stigmer.Run()`)
  - `HttpCallTask()` with `WithHTTPGet()` (should use `wf.HttpGet()`)
  - `FieldRef()` (should use `task.Field()`)
  - `.ThenRef()` (should use implicit dependencies)
- **TODO**: Migrate to new API or mark as legacy

### ⚠️ 09_workflow_with_loops.go
- **Status**: Uses OLD API
- **Features**: FOR tasks, iteration over collections
- **Old patterns**: Similar to above
- **TODO**: Migrate to new API or mark as legacy

### ⚠️ 10_workflow_with_error_handling.go
- **Status**: Uses OLD API
- **Features**: TRY/CATCH tasks, error handling
- **Old patterns**: Similar to above
- **TODO**: Migrate to new API or mark as legacy

### ⚠️ 11_workflow_with_parallel_execution.go
- **Status**: Uses OLD API
- **Features**: FORK tasks, parallel execution
- **Old patterns**: Similar to above
- **TODO**: Migrate to new API or mark as legacy

### ✅ 07_basic_workflow_legacy.go
- **Status**: Intentionally preserved as reference
- **Purpose**: Show OLD API for comparison
- **Use case**: Migration reference

## Migration Strategy

### Option 1: Update All Examples (Recommended)
Migrate examples 08-11 to use the new Pulumi-aligned API patterns:
- Replace `defer stigmer.Complete()` with `stigmer.Run()`
- Replace `HttpCallTask()` with `wf.HttpGet/Post/etc()`
- Replace `FieldRef()` with `task.Field()`
- Remove manual `ThenRef()` (use implicit dependencies)
- Use typed context for configuration

**Benefits**:
- All examples demonstrate best practices
- Users learn one consistent API
- Clear migration path

**Effort**: 2-3 hours (0.5 hour per example)

### Option 2: Mark as Legacy
Rename examples 08-11 to include "_legacy" suffix and keep them as-is for reference.

Create new versions demonstrating advanced features with new API.

**Benefits**:
- Preserves working examples
- Clear OLD vs NEW separation

**Effort**: 4-6 hours (new examples from scratch)

### Option 3: Hybrid Approach
Keep examples 08-11 as-is but add clear headers indicating they use OLD API patterns and link to migration guide.

**Benefits**:
- Minimal effort
- Users can still learn advanced features
- Migration guide helps them translate

**Effort**: 30 minutes (add headers + links)

## Recommendation

**Phase 6 (Current)**: Use Option 3 (Hybrid) - add clear headers to examples 08-11

Add this header to each:

```go
// ⚠️  WARNING: This example uses the OLD API
// This example has not been migrated to the new Pulumi-aligned API.
// It demonstrates [FEATURE] concepts but uses deprecated patterns.
//
// For migration guidance, see: docs/guides/typed-context-migration.md
// For new API patterns, see: examples/07_basic_workflow.go
//
// OLD patterns used:
// - defer stigmer.Complete() → should use stigmer.Run()
// - HttpCallTask() → should use wf.HttpGet/Post/etc()
// - FieldRef() → should use task.Field()
// - .ThenRef() → should use implicit dependencies
```

**Phase 5.3 (Future)**: Execute Option 1 - migrate all examples to new API

This provides:
1. **Immediate clarity** (users know these are old patterns)
2. **Feature documentation** (conditionals, loops, etc. are still useful)
3. **Migration help** (links to guides)
4. **Future path** (can be updated in Phase 5.3)

## Status Summary

| Example | API Version | Status | Action |
|---------|-------------|--------|--------|
| 01_basic_agent.go | N/A (agent) | ✅ Current | None |
| 02_agent_with_skills.go | N/A (agent) | ✅ Current | None |
| 03_agent_with_mcp_servers.go | N/A (agent) | ✅ Current | None |
| 04_agent_with_subagents.go | N/A (agent) | ✅ Current | None |
| 05_agent_with_environment_variables.go | N/A (agent) | ✅ Current | None |
| 06_agent_with_instructions_from_files.go | N/A (agent) | ✅ Current | None |
| 07_basic_workflow.go | NEW | ✅ Updated | None |
| 07_basic_workflow_legacy.go | OLD | ✅ Preserved | None (reference) |
| 08_agent_with_typed_context.go | NEW | ✅ Updated | None |
| 08_workflow_with_conditionals.go | OLD | ⚠️ Needs update | Add header |
| 09_workflow_with_loops.go | OLD | ⚠️ Needs update | Add header |
| 09_workflow_and_agent_shared_context.go | NEW | ✅ Updated | None |
| 10_workflow_with_error_handling.go | OLD | ⚠️ Needs update | Add header |
| 11_workflow_with_parallel_execution.go | OLD | ⚠️ Needs update | Add header |

**Summary**:
- ✅ Updated: 3 workflow examples (07, 09 shared)
- ✅ Current: 6 agent examples (01-06) + 1 agent context (08)
- ⚠️ Need headers: 4 workflow examples (08-11, not including 09 shared)
- ✅ Preserved: 1 legacy example (07_legacy)

## Next Steps

1. Add warning headers to examples 08-11 workflow files
2. Update this README status when examples are migrated
3. Consider Phase 5.3 for full migration

---

*Last Updated: 2026-01-16 Evening (Phase 6 Documentation)*
