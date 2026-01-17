# Learning Log: Stigmer Go SDK Implementation

This log captures patterns, solutions, and gotchas discovered while implementing Stigmer SDK features in Go.

---

## 2026-01-17: Automatic Context Variable Injection

### Context

Implemented Pulumi-style automatic context variable injection - users define variables via `ctx.SetX()` and they're automatically available in workflows.

### Pattern 1: ToValue() Interface for Serialization

**Problem**: Need to extract values from typed Refs (StringRef, IntRef, BoolRef, ObjectRef) for synthesis without losing type information.

**Solution**: Add interface method that each type implements:

```go
// In Ref interface
type Ref interface {
    Expression() string
    Name() string
    IsSecret() bool
    ToValue() interface{}  // ← Returns value for serialization
}

// Each typed Ref implements it
func (s *StringRef) ToValue() interface{} { return s.value }
func (i *IntRef) ToValue() interface{} { return i.value }
func (b *BoolRef) ToValue() interface{} { return b.value }
func (o *ObjectRef) ToValue() interface{} { return o.value }
```

**Why This Works**:
- Type-safe at SDK level (StringRef.Value() returns string)
- JSON-compatible at synthesis (ToValue() returns interface{})
- Extensible for future Ref types
- Clean separation of concerns

**When to Use**: Anytime you need to serialize SDK types to protobuf/JSON

### Pattern 2: Automatic Task Injection at Synthesis

**Problem**: Want variables to "just work" without manual initialization - Pulumi-style DX.

**Solution**: Inject initialization task automatically during synthesis:

```go
func workflowSpecToProtoWithContext(wf *workflow.Workflow, contextVars map[string]interface{}) {
    spec := &workflowv1.WorkflowSpec{...}
    
    // Inject context initialization FIRST
    if len(contextVars) > 0 {
        initTask, _ := createContextInitTask(contextVars)
        spec.Tasks = append(spec.Tasks, initTask)  // ← First task
    }
    
    // Then add user tasks
    for _, task := range wf.Tasks {
        spec.Tasks = append(spec.Tasks, task)
    }
}
```

**Why This Works**:
- User doesn't think about plumbing
- Variables initialized before any user code runs
- Clean separation: SDK handles infrastructure, user writes logic

**When to Use**: Anytime you need to inject setup/teardown tasks automatically

### Pattern 3: Type Serialization to Protobuf

**Problem**: Go types (string, int, bool, map) need to serialize to protobuf Value types correctly.

**Solution**: Use google.protobuf.Struct with proper type mapping:

```go
variables := make(map[string]interface{})
for name, ref := range contextVars {
    variables[name] = ref.ToValue()  // Extract typed value
}

// Convert to protobuf Struct
taskConfig, _ := structpb.NewStruct(map[string]interface{}{
    "variables": variables,
})
```

**Result**:
- `string` → `string_value` in proto
- `int` → `number_value` in proto (JSON numbers are float64)
- `bool` → `bool_value` in proto
- `map[string]interface{}` → `struct_value` in proto (nested)

**Gotcha**: JSON numbers are always float64, even for integers! Protobuf handles the conversion correctly.

**When to Use**: Anytime you're converting Go maps to protobuf Structs

### Pattern 4: Testing Protobuf Output

**Problem**: Need to verify synthesized protobuf manifest has correct structure and values.

**Solution**: Parse and inspect programmatically:

```go
// 1. Generate manifest to file
err := stigmer.Run(func(ctx *stigmer.Context) error {
    // ... SDK code ...
})

// 2. Read generated protobuf
data, _ := os.ReadFile(outputDir + "/workflow-manifest.pb")
manifest := &workflowv1.WorkflowManifest{}
proto.Unmarshal(data, manifest)

// 3. Inspect structure
initTask := manifest.Workflows[0].Spec.Tasks[0]
assert.Equal(t, "__stigmer_init_context", initTask.Name)
assert.Equal(t, "WORKFLOW_TASK_KIND_SET", initTask.Kind.String())

// 4. Verify variables
varsStruct := initTask.TaskConfig.Fields["variables"].GetStructValue()
assert.Equal(t, "https://api.example.com", 
    varsStruct.Fields["apiURL"].GetStringValue())
```

**Why This Works**:
- End-to-end verification (synthesis → proto → parse)
- Catches type conversion issues
- Verifies actual runtime structure

**When to Use**: Integration tests for synthesis features

### Architectural Decision: Internal Variables vs External Config

**Confusion Clarified**:

During implementation, there was confusion about when to use:
1. SET task injection (what we built)
2. ExecutionContext + env_spec (future feature)

**The Distinction**:

| Feature | ctx.SetX (Internal) | ctx.Env (External) |
|---------|--------------------|--------------------|
| **Purpose** | Workflow logic, constants | Secrets, runtime config |
| **Set When** | Synthesis time (hardcoded) | Execution time (injected) |
| **Storage** | Workflow YAML (SET task) | ExecutionContext (encrypted) |
| **Equivalent** | N/A (Pulumi has no direct match) | Pulumi's config.Get() |

**Rule of Thumb**:
- **Use SET task** (ctx.SetX) for: Loop counters, API URLs, retry counts, default values
- **Use env_spec** (ctx.Env - future) for: API keys, database passwords, environment-specific config

**Why It Matters**: Using the wrong approach leads to security issues (secrets in YAML) or complexity (hardcoded values that should be configurable).

---

## 2026-01-17: Auto-Export on Field Reference (Pulumi-Style Implicit Dependencies)

### Context

Implemented automatic export when task output fields are referenced via `.Field()` method. This achieves Pulumi-style implicit dependencies where accessing an output automatically makes it available, eliminating manual `.ExportAll()` calls and preventing "nil reference" runtime errors.

### Pattern 1: In-Place Modification for Auto-Export

**Problem**: Users had to manually call `task.ExportAll()` before using `task.Field()` references, creating boilerplate and potential runtime errors if forgotten.

**Solution**: Modify task export in-place when `.Field()` is called:

```go
// In workflow/task.go
func (t *Task) Field(fieldName string) TaskFieldRef {
    // Auto-export: When a task's field is referenced, automatically export the task
    if t.ExportAs == "" {
        t.ExportAs = "${.}"
    }
    
    return TaskFieldRef{
        taskName:  t.Name,
        fieldName: fieldName,
    }
}
```

**Why This Works**:
- Task is a pointer type (`*Task`) - modifications affect the original
- Export is set at point of use (discoverable)
- No separate tracking mechanism needed
- Simpler than visitor pattern or post-processing

**When to Use**: Anytime you need to trigger side effects based on method calls (accessing outputs, referencing resources)

### Pattern 2: Idempotency Check for Safety

**Problem**: Multiple `.Field()` calls on the same task, or calling `.Field()` after `.ExportField()`, could overwrite custom exports.

**Solution**: Check before setting export:

```go
if t.ExportAs == "" {
    t.ExportAs = "${.}"
}
```

**Why This Works**:
- Empty string check is simple and fast
- First `.Field()` call sets export
- Subsequent calls are no-ops
- Custom exports (`.ExportField("specific")`) are preserved

**Benefits**:
- Multiple `.Field()` calls are safe
- User can still override export if needed
- Backward compatible with existing code

**When to Use**: Anytime you're doing automatic behavior that users might customize

### Pattern 3: Export/Reference Alignment Understanding

**Problem**: Understanding how export transforms task output and how field references access it.

**Solution**: Know the transformation:

1. **Export `{ as: '${.}' }`**:
   - Takes current task output (`.`)
   - Makes it available at `$context.<taskName>`
   - For task `fetchData`: output → `$context.fetchData`

2. **Field Reference `fetchTask.Field("title")`**:
   - Generates: `${ $context.fetchData.title }`
   - Reads from: `$context.fetchData.title`

3. **Alignment**:
   ```
   Export:    $context.fetchData        ← Task output stored here
                      ↓
   Reference: $context.fetchData.title  ← Field reads from here
   ```

**Why This Matters**:
- Export and reference must use the same base path
- `${.}` exports to task name automatically (Zigflow behavior)
- Field references assume this structure
- Any mismatch causes runtime "nil reference" errors

**Gotcha**: If you use custom export like `ExportField("specificField")`, it changes the path to `$context.specificField` (not `$context.taskName.specificField`). Field references won't work with this pattern!

**When to Use**: Designing any auto-export or reference system

### Pattern 4: Testing Examples with examples_test.go

**Problem**: Need to verify examples work end-to-end and synthesize correctly.

**Solution**: Follow existing `examples_test.go` pattern:

```go
func TestExample14_AutoExportVerification(t *testing.T) {
    runExampleTest(t, "14_auto_export_verification.go", func(t *testing.T, outputDir string) {
        // 1. Read manifest
        manifestPath := filepath.Join(outputDir, "workflow-manifest.pb")
        assertFileExists(t, manifestPath)
        
        var manifest workflowv1.WorkflowManifest
        readProtoManifest(t, manifestPath, &manifest)
        
        // 2. Verify structure
        workflow := manifest.Workflows[0]
        
        // 3. Find specific task
        var fetchDataTask *workflowv1.WorkflowTask
        for _, task := range workflow.Spec.Tasks {
            if task.Name == "fetchData" {
                fetchDataTask = task
                break
            }
        }
        
        // 4. Verify export is set
        if fetchDataTask.Export == nil {
            t.Error("fetchData should have auto-export")
        } else if fetchDataTask.Export.As != "${.}" {
            t.Errorf("export.as = %v, want ${.}", fetchDataTask.Export.As)
        }
    })
}
```

**Why This Works**:
- `runExampleTest()` helper handles execution with `STIGMER_OUT_DIR`
- Tests actual synthesis (not mocked)
- Verifies protobuf structure
- Follows established pattern (consistent with other 14 tests)

**Pattern Elements**:
- Test function: `TestExample##_Description`
- Use `runExampleTest()` helper
- Verify manifest file created
- Unmarshal and validate proto
- Check specific fields and values

**When to Use**: Every new example should have a test in `examples_test.go`

### UX Decision: Auto-Export vs Manual Export

**Design Choice**: Auto-export when `.Field()` is called instead of requiring manual `.ExportAll()`

**Rationale**:
1. **Pulumi Alignment**: In Pulumi, accessing `resource.output` creates implicit dependency - no manual export needed
2. **Pit of Success**: Doing the right thing is automatic, not optional
3. **Error Prevention**: Forgetting `.ExportAll()` caused runtime "nil reference" errors
4. **Code Clarity**: One fewer line, clearer intent

**Alternative Considered**: Require manual `.ExportAll()` and provide clear error messages

**Why Auto-Export Won**:
- Eliminates entire class of errors
- Matches proven Pulumi UX pattern
- More discoverable (just use `.Field()` naturally)
- Backward compatible (manual `.ExportAll()` still works)

**Trade-off Accepted**: Tasks auto-export even if only one field is used (slight overhead, but negligible)

### Testing Strategy: Three Levels

**Problem**: How to verify auto-export works correctly at all levels.

**Solution**: Test at three levels:

1. **Unit Tests** (`workflow/task_test.go`):
   - Test behavior in isolation
   - Edge cases: idempotency, custom exports
   - Fast, focused tests
   - **Issue**: Other existing tests had compilation errors (unrelated)

2. **Integration Tests** (Example files as programs):
   - Real workflow synthesis
   - Export/reference alignment
   - Run with `go run`
   - Living documentation
   - **Result**: Created Examples 14 & 15, both work perfectly

3. **Test Suite** (`examples/examples_test.go`):
   - End-to-end verification
   - Manifest structure validation
   - Proto field verification
   - **Result**: All 12 tests passing

**Why Three Levels**:
- Unit tests catch logic errors
- Integration tests catch synthesis errors
- Test suite catches manifest structure errors
- Different levels find different bugs

**When to Use**: Any significant SDK feature needs all three levels

---

## Future Patterns to Document

- Environment spec implementation (ctx.Env)
- Secret variants for all types (SetSecretInt, SetSecretBool)
- Agent synthesis patterns
- Skill integration patterns

---

*This log grows with each feature implementation. Add entries as you discover new patterns!*
