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

## Future Patterns to Document

- Environment spec implementation (ctx.Env)
- Secret variants for all types (SetSecretInt, SetSecretBool)
- Agent synthesis patterns
- Skill integration patterns

---

*This log grows with each feature implementation. Add entries as you discover new patterns!*
