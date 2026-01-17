# Compile-Time Variable Resolution

**Date**: January 17, 2026  
**Status**: ✅ Implemented

---

## Summary

The Stigmer SDK now resolves context variables at **compile-time** (during synthesis) instead of **runtime** (during workflow execution). This eliminates the need for the `__stigmer_init_context` SET task and improves workflow performance.

## The Change

### Before (Runtime Resolution)

```go
ctx.SetString("apiURL", "https://api.example.com")

// Generated workflow had:
// 1. __stigmer_init_context SET task with {"apiURL": "https://api.example.com"}
// 2. User tasks with ${$context.apiURL} JQ expressions
// 3. Workflow runner resolved variables at execution time
```

**Generated Manifest:**
```yaml
tasks:
  - name: __stigmer_init_context
    kind: SET
    task_config:
      variables:
        apiURL: "https://api.example.com"
    export:
      as: "${.}"
  
  - name: fetch_data
    kind: HTTP_CALL
    task_config:
      endpoint:
        uri: "${ $context.apiURL }/users"  # Resolved at RUNTIME
```

### After (Compile-Time Resolution)

```go
ctx.SetString("apiURL", "https://api.example.com")

// Generated workflow has:
// 1. NO __stigmer_init_context task
// 2. User tasks with actual values baked in
// 3. Variables resolved ONCE at synthesis time
```

**Generated Manifest:**
```yaml
tasks:
  - name: fetch_data
    kind: HTTP_CALL
    task_config:
      endpoint:
        uri: "https://api.example.com/users"  # Resolved at SYNTHESIS TIME
```

---

## How It Works

### 1. Variable Storage (Context)

When you call `ctx.SetString()`, `ctx.SetInt()`, etc., the SDK stores the value in memory:

```go
// stigmer/context.go
func (c *Context) SetString(name, value string) *StringRef {
    ref := &StringRef{
        baseRef: baseRef{name: name},
        value:   value,  // Stored for synthesis
    }
    c.variables[name] = ref
    return ref
}
```

### 2. Variable Interpolation (Synthesis)

During synthesis, the SDK finds all `${variableName}` placeholders in task configurations and replaces them with actual values:

```go
// internal/synth/interpolator.go
func InterpolateVariables(taskConfig, contextVars) {
    // Find "${variableName}" and replace with value
    // Preserves types: numbers stay numbers, bools stay bools
}
```

### 3. Manifest Generation (Workflow Converter)

The workflow converter applies interpolation to each task before converting to protobuf:

```go
// internal/synth/workflow_converter.go
func taskToProtoWithInterpolation(task, contextVars) {
    taskConfig := taskConfigToStruct(task)
    
    // Apply variable interpolation
    interpolatedConfig := InterpolateVariables(taskConfig.AsMap(), contextVars)
    
    // Convert to protobuf with resolved values
    return protoTask
}
```

---

## Interpolation Rules

The interpolator handles two cases:

### Case 1: Complete Value Replacement

When a placeholder is the **entire** value, type is preserved:

```json
Input:  {"retries": "${maxRetries}", "enabled": "${isProd}"}
Values: {"maxRetries": 3, "isProd": true}
Output: {"retries": 3, "enabled": true}  // Numbers and bools, not strings!
```

### Case 2: Partial String Replacement

When a placeholder is **part of** a string, it's unwrapped and concatenated:

```json
Input:  {"url": "${baseURL}/api/v${version}/users"}
Values: {"baseURL": "https://api.example.com", "version": "1"}
Output: {"url": "https://api.example.com/api/v1/users"}
```

### Case 3: Complex Types

Objects and arrays are inserted as-is:

```json
Input:  {"config": "${dbConfig}"}
Values: {"dbConfig": {"host": "localhost", "port": 5432}}
Output: {"config": {"host": "localhost", "port": 5432}}
```

---

## Benefits

### 1. Simpler Workflow Manifests

**Before**: 1 SET task + N user tasks  
**After**: N user tasks only

### 2. Faster Execution

- Variables resolved **once** at synthesis time
- No runtime variable resolution overhead
- Workflow runner has less work to do

### 3. Easier Debugging

You can see the actual values in the generated manifest:

```yaml
# Before
uri: "${ $context.apiURL }/users"  # What's the value? 🤷

# After  
uri: "https://api.example.com/users"  # Clear! ✅
```

### 4. Better Performance

- No JQ expression parsing at runtime
- No context lookup overhead
- Fewer tasks to execute

---

## Migration Guide

### For SDK Users

**No changes required!** Your existing code works exactly the same:

```go
// This still works
ctx.SetString("apiURL", "https://api.example.com")
ctx.SetInt("retries", 3)
ctx.SetBool("isProd", true)

// Use variables in task configs
// "${apiURL}/users" → "https://api.example.com/users" (compile-time)
```

### For Workflow Runner

The generated manifests no longer include `__stigmer_init_context` tasks. If your runner has special handling for this task, it can be removed.

---

## Example Usage

```go
package main

import (
    "github.com/leftbin/stigmer-sdk/go/stigmer"
    "github.com/leftbin/stigmer-sdk/go/workflow"
)

func main() {
    stigmer.Run(func(ctx *stigmer.Context) error {
        // Define variables (stored in context)
        apiURL := ctx.SetString("apiURL", "https://api.example.com")
        apiKey := ctx.SetSecret("apiKey", "secret-key-123")
        retries := ctx.SetInt("retries", 3)
        timeout := ctx.SetInt("timeout", 30)
        
        // Create workflow
        wf, _ := workflow.New(ctx,
            workflow.WithName("api-client"),
            workflow.WithNamespace("examples"),
        )
        
        // Use variables - they'll be resolved at synthesis time!
        // "${apiURL}/users" → "https://api.example.com/users"
        // "${retries}" → 3 (number, not string)
        
        return nil
    })
}
```

**Generated Manifest** (simplified):
```yaml
workflows:
  - name: api-client
    namespace: examples
    tasks: []  # Your tasks with values already baked in!
              # NO __stigmer_init_context task!
```

---

## Runtime vs Compile-Time Variables

### Compile-Time Variables (New Default)

- **When**: Resolved during `stigmer.Run()` → `ctx.Synthesize()`
- **How**: `${variableName}` placeholders replaced with values
- **Use**: Configuration, URLs, secrets, retry counts
- **Example**: `ctx.SetString("apiURL", "https://api.example.com")`

### Runtime Variables (Task Outputs)

- **When**: Resolved during workflow execution
- **How**: JQ expressions like `${ $context.taskName.field }`
- **Use**: Task outputs, dynamic data from API responses
- **Example**: `${ $context.fetchUser.id }`

**You can mix both!**

```yaml
task_config:
  endpoint:
    uri: "https://api.example.com/users/${.user_id}"
        #     ^^^^^^^^^^^^^^^^^^^^^ compile-time
        #                                   ^^^^^^^^^ runtime (from previous task)
```

---

## Implementation Details

### Files Changed

1. **`internal/synth/interpolator.go`** (new)
   - `InterpolateVariables()`: Main interpolation logic
   - `replaceVariablePlaceholders()`: Two-pass replacement (complete values, then partials)

2. **`internal/synth/workflow_converter.go`** (modified)
   - `workflowSpecToProtoWithContext()`: No longer generates SET task
   - `taskToProtoWithInterpolation()`: Applies interpolation to each task
   - `createContextInitTask()`: Deprecated (kept for reference)

3. **`stigmer/context.go`** (docs updated)
   - Updated comments to reflect compile-time resolution

4. **`stigmer/refs.go`** (docs updated)
   - Clarified distinction between compile-time and runtime resolution

### Tests

**`internal/synth/interpolator_test.go`**: Comprehensive tests covering:
- String, int, bool, object, array interpolation
- Complete value vs partial string replacement
- Nested objects and special characters
- Missing variable error handling

All tests pass ✅

---

## Conceptual Shift

This change aligns the SDK with infrastructure-as-code tools like Pulumi and Terraform:

| Tool | Variable Resolution |
|------|---------------------|
| **Pulumi** | Compile-time (during `pulumi up`) |
| **Terraform** | Compile-time (during `terraform apply`) |
| **Stigmer SDK (old)** | Runtime (during workflow execution) ❌ |
| **Stigmer SDK (new)** | Compile-time (during synthesis) ✅ |

**Why this matters:**
- Infrastructure tools resolve configuration **before** executing
- This makes manifests deterministic and debuggable
- Runtime resolution is reserved for **dynamic data** (API responses, task outputs)

---

## Limitations

### 1. Static Values Only

Context variables must be known at synthesis time:

```go
// ✅ Works (static value)
ctx.SetString("apiURL", "https://api.example.com")

// ❌ Won't work (runtime value)
ctx.SetString("apiURL", os.Getenv("API_URL"))  // Environment var might change!
```

**Workaround**: Read environment variables during `stigmer.Run()`:

```go
stigmer.Run(func(ctx *stigmer.Context) error {
    apiURL := os.Getenv("API_URL")  // Read once at synthesis time
    ctx.SetString("apiURL", apiURL)  // Store as compile-time constant
    // ...
})
```

### 2. No Circular References

Variables can't reference each other:

```go
// ❌ Not supported
baseURL := ctx.SetString("baseURL", "https://api.example.com")
fullURL := ctx.SetString("fullURL", "${baseURL}/users")  // Won't interpolate!
```

**Workaround**: Use string operations in Go:

```go
// ✅ Works
baseURL := ctx.SetString("baseURL", "https://api.example.com")
fullURL := ctx.SetString("fullURL", baseURL.Value() + "/users")
```

---

## Future Enhancements

### Potential Improvements

1. **Interpolation for Runtime Expressions**: Support `${baseURL}` in JQ expressions
2. **Variable Dependencies**: Allow variables to reference other variables
3. **Secret Masking**: Hide secret values in debug output
4. **Variable Validation**: Type checking and required variable enforcement

---

## Questions?

For more information:
- See `internal/synth/interpolator.go` for implementation
- See `internal/synth/interpolator_test.go` for examples
- Check `stigmer/context.go` for context variable API

---

**Implementation**: Suresh  
**Review**: Stigmer Team  
**Status**: ✅ Shipped in SDK v0.2.0
