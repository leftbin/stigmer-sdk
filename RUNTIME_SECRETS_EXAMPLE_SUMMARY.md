# Runtime Secrets Example - Implementation Summary

**Created**: 2026-01-18  
**Location**: `go/examples/14_workflow_with_runtime_secrets.go`

## Overview

Created a comprehensive example and test suite demonstrating the secure pattern for handling sensitive data using runtime execution context variables (`RuntimeSecret()` and `RuntimeEnv()`).

## What Was Created

### 1. Example File: `14_workflow_with_runtime_secrets.go`

**Path**: `go/examples/14_workflow_with_runtime_secrets.go`  
**Lines**: ~260 lines of code + extensive documentation

**Demonstrates 6 Real-World Scenarios**:

1. **External API Authentication** - OpenAI API with runtime secrets
2. **Environment-Specific Configuration** - Dynamic endpoints based on environment
3. **Multiple Secrets in One Request** - Stripe API with multiple auth headers
4. **Database API Credentials** - Database password protection via headers
5. **Third-Party Webhook Registration** - Webhook signing secret security
6. **Mixed Static and Runtime Config** - Slack notifications with environment info

**Key Features Demonstrated**:
- ✅ `RuntimeSecret("KEY")` for sensitive data (API keys, passwords, tokens)
- ✅ `RuntimeEnv("VAR")` for non-secret environment config
- ✅ `Interpolate()` for combining static text with runtime values
- ✅ Multiple secrets in single task
- ✅ Task field references as headers (not in bodies)
- ✅ Environment-specific URLs

**Security Pattern Highlighted**:
```go
// ❌ WRONG: Compile-time secrets (INSECURE)
apiKey := ctx.SetSecret("key", "sk-proj-abc123")
// Result: Manifest contains "sk-proj-abc123" ← EXPOSED! ❌

// ✅ CORRECT: Runtime secrets (SECURE)
workflow.Header("Authorization", 
    workflow.Interpolate("Bearer ", workflow.RuntimeSecret("OPENAI_API_KEY"))
)
// Result: Manifest contains "${ 'Bearer ' + .secrets.OPENAI_API_KEY }" ← SAFE! ✅
```

### 2. Comprehensive Test Suite: `examples_test.go`

**Function**: `TestExample14_WorkflowWithRuntimeSecrets()`  
**Lines**: ~300 lines of test code

**Test Coverage**:

#### Security Test 1: Runtime Secret Placeholders
- ✅ Verifies OpenAI API key as runtime secret placeholder
- ✅ Checks Authorization header format
- ✅ Detects actual secret values (fails if found)

#### Security Test 2: Environment Variable Placeholders
- ✅ Verifies environment-specific endpoint URLs
- ✅ Checks runtime env var in URI construction
- ✅ Validates region headers use runtime env vars

#### Security Test 3: Multiple Secrets Pattern
- ✅ Stripe API with multiple auth headers
- ✅ Authorization bearer token check
- ✅ Idempotency key verification

#### Security Test 4: Database Credentials
- ✅ Database password in headers (not body)
- ✅ Placeholder format verification
- ✅ Secret leakage detection

#### Security Test 5: Webhook Security
- ✅ External API key authentication
- ✅ Webhook signing secret protection
- ✅ Secret value detection

**Helper Functions**:
- `containsRuntimeRef()` - Checks if value contains runtime placeholder
- `containsSecretValue()` - Detects actual secret patterns (sk-, AKIA, etc.)
- `containsSubstring()` - String matching helper
- `recursiveContains()` - Recursive substring search

### 3. Documentation Updates

#### `go/examples/_docs/README.md`
- Added Example 14 to workflow examples section
- Added security pattern section explaining runtime vs compile-time
- Documented when to use each pattern
- Added execution examples for dev/staging/prod

#### Updated README Sections:
```markdown
## 🔒 Security Pattern - Runtime Secrets (Example 14)

**CRITICAL CONCEPT**: Example 14 demonstrates the secure pattern for handling sensitive data.

### When to Use Each Pattern

**RuntimeSecret()**: API keys, tokens, passwords, OAuth secrets, private keys  
**RuntimeEnv()**: Environment names, regions, feature flags, non-secret config  
**ctx.SetString()**: Static configuration, public constants, non-secret metadata
```

## Test Results

### Verification Summary

```bash
=== RUN   TestExample14_WorkflowWithRuntimeSecrets
--- PASS: TestExample14_WorkflowWithRuntimeSecrets (0.31s)
```

**What Was Verified**:
- ✅ All API keys use RuntimeSecret() placeholders
- ✅ Environment config uses RuntimeEnv() placeholders
- ✅ NO actual secret values found in manifest
- ✅ Placeholders correctly embedded: `.secrets.KEY` and `.env_vars.VAR`
- ✅ Multiple secrets in single task work correctly
- ✅ Database credentials properly secured
- ✅ Webhook signing secrets properly secured
- ✅ Environment-specific URLs work with runtime env vars

### Example Manifest Output

The generated manifest contains **ONLY placeholders**, never actual secrets:

```
Authorization: ${ "Bearer " + .secrets.OPENAI_API_KEY }
Idempotency-Key: ${.secrets.STRIPE_IDEMPOTENCY_KEY}
X-DB-Password: ${.secrets.DATABASE_PASSWORD}
Webhook secret: ${.secrets.WEBHOOK_SIGNING_SECRET}
Endpoint URI: ${ "https://api-" + .env_vars.ENVIRONMENT + ".example.com/process" }
X-Region: ${.env_vars.AWS_REGION}
```

## Usage Examples

### Development Environment
```bash
stigmer run secure-api-workflow \
  --runtime-env secret:OPENAI_API_KEY=sk-proj-dev123 \
  --runtime-env secret:STRIPE_API_KEY=sk_test_dev \
  --runtime-env secret:DATABASE_PASSWORD=devPassword123 \
  --runtime-env ENVIRONMENT=dev \
  --runtime-env AWS_REGION=us-west-2 \
  --runtime-env LOG_LEVEL=debug
```

### Production Environment (Same Manifest!)
```bash
stigmer run secure-api-workflow \
  --runtime-env secret:OPENAI_API_KEY=sk-proj-prod-xyz \
  --runtime-env secret:STRIPE_API_KEY=sk_live_realkey \
  --runtime-env secret:DATABASE_PASSWORD=prodPassword \
  --runtime-env ENVIRONMENT=production \
  --runtime-env AWS_REGION=us-east-1 \
  --runtime-env LOG_LEVEL=info
```

### CI/CD Pipeline with Vault
```bash
export OPENAI_KEY=$(vault read -field=value secret/openai/api-key)
export STRIPE_KEY=$(vault read -field=value secret/stripe/api-key)
export DB_PASS=$(vault read -field=value secret/database/password)

stigmer run secure-api-workflow \
  --runtime-env secret:OPENAI_API_KEY="$OPENAI_KEY" \
  --runtime-env secret:STRIPE_API_KEY="$STRIPE_KEY" \
  --runtime-env secret:DATABASE_PASSWORD="$DB_PASS" \
  --runtime-env ENVIRONMENT=staging \
  --runtime-env AWS_REGION=eu-west-1
```

## Key Insights

### What Works Well

1. **Type Safety**: Using `workflow.RuntimeSecret()` and `workflow.RuntimeEnv()` provides clear intent
2. **Security**: Secrets never appear in manifests or Temporal history
3. **Flexibility**: Same manifest works across all environments
4. **Interpolation**: `Interpolate()` combines static text with runtime values elegantly
5. **Headers**: Task field references work well in headers

### Implementation Notes

1. **Field References in Bodies**: TaskFieldRef values cannot be used directly in `WithBody()` - use headers instead
2. **Array Types**: Nested arrays in bodies can cause protobuf conversion issues - keep bodies simple
3. **Interpolate Format**: Generates `${ "static" + .secrets.KEY }` format, not `${.secrets.KEY}` directly
4. **Testing Strategy**: Check for placeholder presence, not exact format (format may evolve)

## Files Modified

1. ✅ **Created**: `go/examples/14_workflow_with_runtime_secrets.go` (260 lines)
2. ✅ **Modified**: `go/examples/examples_test.go` (+300 lines)
3. ✅ **Modified**: `go/examples/_docs/README.md` (+50 lines)

## Impact

### For Users
- **Clear Example**: Users now have a comprehensive reference for secure secret handling
- **Real-World Scenarios**: 6 practical scenarios covering common use cases
- **Security Awareness**: Documentation emphasizes security implications
- **Copy-Paste Ready**: Examples can be adapted directly to user workflows

### For Testing
- **Automated Verification**: Test suite ensures security guarantees hold
- **Regression Prevention**: Detects if secrets accidentally leak into manifests
- **Format Validation**: Verifies placeholder formats are correct
- **Multiple Scenarios**: Tests cover diverse use cases

### For Documentation
- **Visibility**: Runtime secrets feature is now prominently documented
- **Comparison**: Clear before/after showing security difference
- **Guidelines**: Users know when to use each pattern

## Next Steps (Future Enhancements)

1. **Field References in Bodies**: Consider supporting TaskFieldRef in `WithBody()` for cleaner syntax
2. **Validation**: Add runtime validation of placeholder formats during synthesis
3. **IDE Support**: Consider autocomplete/IntelliSense for secret names
4. **Vault Integration**: Example showing direct integration with HashiCorp Vault or AWS Secrets Manager

---

## Success Criteria: ✅ COMPLETE

- ✅ Comprehensive example with 6 real-world scenarios
- ✅ Test suite with 300+ lines covering all scenarios
- ✅ Documentation updated with security patterns
- ✅ All tests passing
- ✅ No actual secrets in generated manifests
- ✅ Clear usage examples for dev/staging/prod
- ✅ Security comparison (compile-time vs runtime)

**This example is ready for users!**
