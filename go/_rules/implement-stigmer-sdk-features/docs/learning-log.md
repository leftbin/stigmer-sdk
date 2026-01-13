# Stigmer Go SDK Learning Log

## Purpose

This log captures all learnings, discoveries, and solutions from implementing and maintaining the Stigmer Go SDK. Organized by topic for easy lookup.

**Check this log BEFORE solving a problem** - the solution might already be documented!

---

## How to Use This Log

1. **Before implementing**: Search for your topic/error
2. **Found a solution**: Check if it's already documented
3. **New discovery**: Add it to the appropriate section
4. **Organized by topic**: Not chronological, easier to find

---

## Proto Converters & Transformations

**Topic Coverage**: Proto message conversions, pointer handling, nil checks, nested messages, repeated fields

### 2026-01-13 - Proto-Agnostic SDK Architecture

**Problem**: SDK was tightly coupled to proto definitions. All packages had `ToProto()` methods and proto imports, making the SDK difficult to evolve independently and exposing proto complexity to users.

**Root Cause**: 
- Original design assumed SDK should handle proto conversion
- Mixing concerns: user API (SDK) and platform communication (proto)
- Proto changes would require SDK API changes
- Users had to understand proto to use SDK

**Solution**: Make SDK completely proto-agnostic - remove all proto dependencies

**Implementation in Go**:

```go
// ❌ BEFORE: SDK exposed proto
package agent

import (
    agentv1 "github.com/leftbin/stigmer/apis/stubs/go/ai/stigmer/agentic/agent/v1"
)

func (a *Agent) ToProto() *agentv1.AgentSpec {
    // SDK handles proto conversion
}

// ✅ AFTER: SDK is proto-agnostic
package agent

import (
    "os"  // No proto imports!
)

// SDK only defines user-friendly Go structs
// CLI handles all proto conversion
```

**Benefits**:
- SDK evolves independently from proto
- Cleaner API surface (no proto exposure)
- Better separation of concerns (SDK = user API, CLI = proto converter)
- Users never see proto complexity
- Can change proto without breaking SDK API

**Architecture Pattern**:
```
Before: SDK -> Proto -> CLI -> Platform
After:  SDK (proto-agnostic) -> CLI (proto converter) -> Platform
```

**Prevention**: 
- Never import proto packages in SDK code
- SDK provides Go structs and interfaces
- CLI reads SDK objects and converts to proto
- Keep SDK user-focused, not protocol-focused

**Cross-Language Reference**: Python SDK had similar coupling - both SDKs benefit from proto-agnostic design

---

### 2026-01-13 - Removing ToProto() Methods

**Problem**: `ToProto()` methods in SDK packages created tight coupling to proto definitions and made testing harder.

**Root Cause**: 
- Assumed SDK should do proto conversion
- Mixed serialization logic with domain logic
- Required proto stubs in SDK dependencies

**Solution**: Remove all `ToProto()` methods from SDK packages

**Files Changed**:
- `agent/agent.go` - Removed `ToProto()` and `convertSkills/MCPServers/SubAgents()`
- `skill/skill.go` - Removed `ToProto()`
- `environment/environment.go` - Removed `ToProto()` and `ToEnvironmentSpec()`
- `subagent/subagent.go` - Removed `ToProto()`
- `mcpserver/*.go` - Removed `ToProto()` from all server types

**Result**:
```go
// Users just define agents in Go
agent, _ := agent.New(
    agent.WithName("reviewer"),
    agent.WithInstructions("..."),
)

// No ToProto() calls needed
// CLI handles everything: stigmer deploy agent.go
```

**Prevention**: 
- SDK should only define domain objects
- CLI is responsible for proto conversion
- Keep serialization separate from domain logic

---

## Agent Configuration & Setup

**Topic Coverage**: Builder patterns, struct composition, configuration validation, environment setup

### 2026-01-13 - File-Based Content Loading

**Problem**: Instructions and skill markdown are often hundreds of lines long. Embedding them as Go strings is impractical and makes code hard to read.

**Root Cause**:
- Long text doesn't belong inline in code
- Version control diffs become unreadable
- Doesn't match IaC patterns (Pulumi, Terraform)

**Solution**: Add file-loading functions to read content from `.md` files

**Implementation in Go**:

```go
// For agent instructions
func WithInstructionsFromFile(path string) Option {
    return func(a *Agent) error {
        content, err := os.ReadFile(path)
        if err != nil {
            return fmt.Errorf("failed to read instructions: %w", err)
        }
        a.Instructions = string(content)
        return nil
    }
}

// For skill markdown
func WithMarkdownFromFile(path string) Option {
    return func(s *Skill) error {
        content, err := os.ReadFile(path)
        if err != nil {
            return fmt.Errorf("failed to read markdown: %w", err)
        }
        s.MarkdownContent = string(content)
        return nil
    }
}
```

**Usage Pattern**:

```go
// Clean repository structure
// my-agent-repo/
// ├── agent.go
// ├── instructions/
// │   └── code-reviewer.md
// └── skills/
//     └── code-analyzer.md

agent, _ := agent.New(
    agent.WithName("code-reviewer"),
    agent.WithInstructionsFromFile("instructions/code-reviewer.md"),
)

skill, _ := skill.New(
    skill.WithName("analyzer"),
    skill.WithMarkdownFromFile("skills/code-analyzer.md"),
)
```

**Benefits**:
- Long content lives in dedicated files
- Better version control (clearer diffs)
- Matches Pulumi/Terraform module patterns
- Users can organize content logically
- Syntax highlighting in editors

**Prevention**: 
- Always provide file-based options for long text
- Support both inline strings and file paths
- Use `os.ReadFile()` for file loading
- Return descriptive errors if file missing

**Cross-Language Reference**: Python SDK can use similar pattern with `Path(...).read_text()`

---

### 2026-01-13 - Inline Resource Creation (Skills)

**Problem**: Users had to pre-create skills on platform before referencing them in agents. This broke the "module authoring" experience.

**Root Cause**:
- SDK only supported references to existing resources
- No way to define custom skills in repository
- Manual platform operations required

**Solution**: Support inline skill creation with `skill.New()`

**Implementation in Go**:

```go
// Skill struct supports both inline and referenced
type Skill struct {
    // For inline skills:
    Name            string // Required for inline
    Description     string // Optional
    MarkdownContent string // Required for inline
    
    // For referenced skills:
    Slug     string // Resource slug
    Org      string // Organization (empty for platform)
    
    // Discriminator:
    IsInline bool
}

// Create inline skill
func New(opts ...Option) (*Skill, error) {
    s := &Skill{IsInline: true}
    for _, opt := range opts {
        if err := opt(s); err != nil {
            return nil, err
        }
    }
    // Validation
    if s.Name == "" {
        return nil, ErrSkillNameRequired
    }
    if s.MarkdownContent == "" {
        return nil, ErrSkillMarkdownRequired
    }
    return s, nil
}

// Reference platform skill
func Platform(slug string) Skill {
    return Skill{
        Slug:     slug,
        IsInline: false,
    }
}

// Reference organization skill
func Organization(org, slug string) Skill {
    return Skill{
        Slug:     slug,
        Org:      org,
        IsInline: false,
    }
}
```

**Usage Pattern**:

```go
// Create inline skill in repository
mySkill, _ := skill.New(
    skill.WithName("code-analyzer"),
    skill.WithDescription("Custom code analysis"),
    skill.WithMarkdownFromFile("skills/analyzer.md"),
)

// Use inline + referenced together
agent, _ := agent.New(
    agent.WithName("reviewer"),
    agent.WithSkills(
        *mySkill,                          // Inline skill
        skill.Platform("security-analysis"), // Platform skill
        skill.Organization("my-org", "internal-docs"), // Org skill
    ),
)
```

**CLI Behavior**:
1. Parse Go code to find inline skills
2. Create inline skills on platform first → get references
3. Convert all skills to `ApiResourceReference`
4. Create agent with skill references

**Benefits**:
- Users define skills in repository (like Pulumi resources)
- No manual pre-creation needed
- CLI orchestrates lifecycle
- ApiResourceReference is deterministic (org + slug known)

**Prevention**: 
- Support both inline and referenced patterns
- Use discriminator field (`IsInline`) to distinguish
- CLI handles creation orchestration
- SDK just defines intent

**Cross-Language Reference**: Python SDK can use same pattern with `Skill(name=..., markdown=...)` vs `Skill.platform(...)`

---

### 2026-01-13 - Builder Pattern Methods

**Problem**: All configuration had to happen in constructor. No way to add components programmatically or conditionally.

**Root Cause**:
- Only functional options at construction time
- No post-creation modification methods
- Inflexible for conditional logic

**Solution**: Add builder methods that return `*Agent` for chaining

**Implementation in Go**:

```go
// Builder methods for post-creation additions
func (a *Agent) AddSkill(s skill.Skill) *Agent {
    a.Skills = append(a.Skills, s)
    return a  // Enable chaining
}

func (a *Agent) AddSkills(skills ...skill.Skill) *Agent {
    a.Skills = append(a.Skills, skills...)
    return a
}

func (a *Agent) AddMCPServer(server mcpserver.MCPServer) *Agent {
    a.MCPServers = append(a.MCPServers, server)
    return a
}

func (a *Agent) AddSubAgent(sub subagent.SubAgent) *Agent {
    a.SubAgents = append(a.SubAgents, sub)
    return a
}

func (a *Agent) AddEnvironmentVariable(variable environment.Variable) *Agent {
    a.EnvironmentVariables = append(a.EnvironmentVariables, variable)
    return a
}
```

**Usage Pattern**:

```go
// Create base agent
agent, _ := agent.New(
    agent.WithName("reviewer"),
    agent.WithInstructions("..."),
)

// Build programmatically
agent.
    AddSkill(mySkill).
    AddSkill(skill.Platform("security")).
    AddMCPServer(githubServer)

// Conditional additions
if needsAnalysis {
    agent.AddSkill(analyzerSkill)
}

// Bulk additions
agent.AddSkills(skill1, skill2, skill3)
```

**Benefits**:
- Flexible composition after creation
- Supports conditional logic
- Method chaining for fluent API
- Can build agents programmatically

**Pattern Details**:
- Methods mutate the receiver (pointer)
- Return `*Agent` to enable chaining
- Use variadic parameters for bulk operations
- Consistent naming: `Add*` for single, `Add*s` for multiple

**Prevention**: 
- Always return `*Agent` from builder methods
- Use pointer receivers for mutation
- Support both single and bulk operations
- Keep functional options for construction-time config

**Cross-Language Reference**: Python can use similar pattern with `agent.add_skill(...)` methods

---

### 2026-01-13 - Completing File Loading Pattern Across Types

**Problem**: File loading support existed for `agent.WithInstructionsFromFile()` and `skill.WithMarkdownFromFile()`, but subagents only supported inline strings. This inconsistency made the API confusing.

**Root Cause**:
- Subagent implementation was added after agent and skill
- File loading wasn't considered for subagents initially
- No systematic check for API consistency across similar types

**Solution**: Add `subagent.WithInstructionsFromFile()` to complete the pattern

**Implementation in Go**:

```go
// Added to subagent package
func WithInstructionsFromFile(path string) InlineOption {
    return func(s *SubAgent) error {
        content, err := os.ReadFile(path)
        if err != nil {
            return err
        }
        s.instructions = string(content)
        return nil
    }
}

// Now all types support file loading consistently:
agent.WithInstructionsFromFile("instructions/agent.md")      // ✓
skill.WithMarkdownFromFile("skills/skill.md")                 // ✓
subagent.WithInstructionsFromFile("instructions/subagent.md") // ✓ NEW
```

**API Consistency Principle**:
When adding a feature to one SDK type, check all similar types:
- **Agent** ← file loading
- **Skill** ← file loading  
- **SubAgent** ← file loading (added)

**Benefits**:
- Consistent API across all SDK types
- Users don't have different patterns for different types
- Predictable behavior
- Easy to remember

**Prevention**: 
- When adding features, audit all similar types
- Maintain feature parity matrix
- Check: agent, skill, subagent, mcpserver, environment
- Document pattern completion in learning log

**Example**:
```go
// All types now support file loading
agent, _ := agent.New(
    agent.WithInstructionsFromFile("instructions/reviewer.md"),
)

inlineSkill, _ := skill.New(
    skill.WithMarkdownFromFile("skills/analyzer.md"),
)

inlineSub, _ := subagent.Inline(
    subagent.WithInstructionsFromFile("instructions/security.md"),
)
```

**Cross-Language Reference**: Python SDK should also check for similar consistency across types

---

## Testing Patterns

**Topic Coverage**: Table-driven tests, test fixtures, mocking, integration tests

### 2026-01-13 - Test Helper Pattern for API Migration

**Problem**: When changing `subagent.Inline()` from non-error to error-returning, all test cases needed updating. Without a helper, test code became verbose with repetitive error handling.

**Root Cause**:
- API signature changed: `func Inline(...) SubAgent` → `func Inline(...) (SubAgent, error)`
- Many test cases use `Inline()` inline in test data
- Can't handle errors in struct literals
- Direct usage would require extracting to variables

**Solution**: Create `mustInline()` test helper that panics on error

**Implementation in Go**:

```go
// Test helper for cleaner test code
func mustInline(opts ...subagent.InlineOption) subagent.SubAgent {
    sub, err := subagent.Inline(opts...)
    if err != nil {
        panic("failed to create inline sub-agent: " + err.Error())
    }
    return sub
}

// Before (error-returning API):
opts: []Option{
    WithName("main-agent"),
    WithSubAgent(mustInline(  // Clean!
        subagent.WithName("helper"),
        subagent.WithInstructions("..."),
    )),
}

// Without helper (verbose):
func() []Option {
    helper, err := subagent.Inline(
        subagent.WithName("helper"),
        subagent.WithInstructions("..."),
    )
    if err != nil {
        t.Fatalf("Failed: %v", err)
    }
    return []Option{
        WithName("main-agent"),
        WithSubAgent(helper),
    }
}()
```

**When to Use This Pattern**:
- Test code only (not production code)
- When API changed from non-error to error-returning
- When tests need inline construction
- When errors indicate test setup problems (not scenarios being tested)

**Pattern Details**:
- Name: `must*` prefix indicates panics on error
- Returns unwrapped value (not error)
- Panics with descriptive message
- Only for test code (helper in `*_test.go`)

**Benefits**:
- Cleaner test code
- Maintains table-driven test structure
- Clear panic messages for debugging
- Follows Go convention (`must*` helpers)

**Prevention**: 
- When migrating APIs to return errors, create `must*` test helpers
- Use only in tests, not production code
- Panic with clear context message
- Document in test file

**Go Convention Reference**: Similar to `template.Must()`, `regexp.MustCompile()` in stdlib

---

## Error Handling

**Topic Coverage**: Error wrapping, custom errors, validation errors, error propagation

### 2026-01-13 - Functional Options Should Return Errors

**Problem**: When adding `WithInstructionsFromFile()` to subagent, the existing functional options didn't support errors. File I/O can fail, but there was no way to propagate the error.

**Root Cause**:
- Original functional options were `func(*SubAgent)` - no error return
- File operations can fail (file not found, permission denied)
- No error propagation path
- Would have to panic or silently ignore errors (both bad)

**Solution**: Change all functional options to return errors

**Implementation in Go**:

```go
// ❌ BEFORE: No error handling
type InlineOption func(*SubAgent)

func WithName(name string) InlineOption {
    return func(s *SubAgent) {
        s.name = name
    }
}

func Inline(opts ...InlineOption) SubAgent {
    s := SubAgent{subAgentType: subAgentTypeInline}
    for _, opt := range opts {
        opt(&s)  // Can't handle errors
    }
    return s
}

// ✅ AFTER: Proper error handling
type InlineOption func(*SubAgent) error

func WithName(name string) InlineOption {
    return func(s *SubAgent) error {
        s.name = name
        return nil  // Non-failing options return nil
    }
}

func WithInstructionsFromFile(path string) InlineOption {
    return func(s *SubAgent) error {
        content, err := os.ReadFile(path)
        if err != nil {
            return err  // File errors propagate
        }
        s.instructions = string(content)
        return nil
    }
}

func Inline(opts ...InlineOption) (SubAgent, error) {
    s := SubAgent{subAgentType: subAgentTypeInline}
    for _, opt := range opts {
        if err := opt(&s); err != nil {
            return SubAgent{}, err  // Propagate errors
        }
    }
    return s, nil
}
```

**Error Handling Pattern**:
1. All options return `error` (even if they never fail)
2. Options that can't fail return `nil`
3. Options that can fail return descriptive errors
4. Constructor checks and propagates errors
5. Users handle errors at construction site

**Usage Pattern**:

```go
// Proper error handling
sub, err := subagent.Inline(
    subagent.WithName("helper"),
    subagent.WithInstructionsFromFile("instructions/helper.md"),
)
if err != nil {
    return fmt.Errorf("creating subagent: %w", err)
}
```

**When to Use Error-Returning Options**:
- ✅ When ANY option might fail (file I/O, validation, parsing)
- ✅ Better to return errors than panic
- ✅ Matches Go idioms (explicit error handling)
- ❌ Don't if absolutely no option can ever fail (rare)

**Benefits**:
- Explicit error handling (Go way)
- Clear error messages
- No hidden panics
- Extensible (can add failing options later)

**Prevention**: 
- Design functional options to return errors from the start
- Even if current options don't fail, future ones might
- Consistent with `agent.Option` and `skill.Option` patterns
- Better safe than refactor later

**Cross-Language Reference**: Python decorators can raise exceptions, similar error propagation model

---

## API Design & Package Organization

**Topic Coverage**: Import cycles, package structure, root package patterns, cross-cutting concerns

### 2026-01-13 - Avoiding Import Cycles with Root Package Pattern

**Problem**: Attempted to add `Complete()` synthesis function to `agent` package, but created an import cycle: `agent` → `synth` (for Complete()) and `synth` → `agent` (for type conversion).

**Root Cause**:
- `agent` package needs to call synthesis logic (`synth`)
- `synth` package needs to convert agent types to proto
- This creates a circular dependency that Go doesn't allow
- Compilation error: "import cycle not allowed"

**Solution**: Move `Complete()` to root package to break the cycle

**Implementation in Go**:

```go
// ❌ BEFORE: Import cycle
// agent/agent.go
package agent

import "github.com/leftbin/stigmer-sdk/go/internal/synth"

func Complete() {
    synth.AutoSynth()  // agent → synth
}

// internal/synth/converter.go
package synth

import "github.com/leftbin/stigmer-sdk/go/agent"

func ToManifest(a *agent.Agent) {...}  // synth → agent
// CYCLE: agent → synth → agent

// ✅ AFTER: Root package breaks cycle
// synthesis.go (root package)
package stigmeragent

import "github.com/leftbin/stigmer-sdk/go/internal/synth"

func Complete() {
    synth.AutoSynth()  // root → synth (no cycle)
}

// internal/synth/converter.go
package synth

import "github.com/leftbin/stigmer-sdk/go/agent"

func ToManifest(a *agent.Agent) {...}  // synth → agent
// NO CYCLE: root → synth, synth → agent, no agent → root
```

**Package Dependency Graph**:

```
Before (CYCLE):
┌──────┐
│agent │ ──→ synth ──→ agent (CYCLE!)
└──────┘       ↑____________│

After (NO CYCLE):
┌──────┐
│ root │ ──→ synth ──→ agent
└──────┘
   ↑                    │
   └─────── user ───────┘
```

**Root Package Pattern**:
- Use root package for **cross-cutting concerns**
- Functions that need to coordinate between packages
- Break import cycles by being "above" all subpackages
- Examples: `Complete()` for synthesis, `Version()` for SDK info

**Usage Pattern**:

```go
import stigmeragent "github.com/leftbin/stigmer-sdk/go"  // Root
import "github.com/leftbin/stigmer-sdk/go/agent"         // Subpackage

func main() {
    defer stigmeragent.Complete()  // Root package function
    
    agent.New(...)  // Subpackage function
}
```

**Benefits**:
- No import cycles
- Clear separation: root = coordination, subpackages = domain
- Users import two packages (predictable pattern)
- Extensible for other cross-cutting concerns

**Prevention**: 
- When adding functions that coordinate between packages, use root package
- Check for cycles: if package A needs package B, and B needs A, use root
- Root package is for SDK-wide operations, not domain logic
- Keep subpackages focused on their domain (agent, skill, etc.)

**Go Import Cycle Detection**:
```bash
# Check for import cycles
go build ./...
# Error will show: "import cycle not allowed in test"
```

---

### 2026-01-13 - Go Language Constraints: No atexit Hooks

**Problem**: User questioned why Go SDK requires `defer stigmeragent.Complete()` when the original design envisioned zero-boilerplate synthesis (like Python's `atexit` hooks).

**Root Cause**:
- **Python has `atexit.register()`** - automatically runs functions on exit
- **TypeScript has `process.on('exit')`** - event-based exit hooks
- **Go has no built-in exit hooks** (until Go 1.24+ with `runtime.AddExitHook`)
- Original design document assumed all languages would have this capability

**Why Go is Different**:

| Language | Exit Hook | Automatic Synthesis? |
|---|---|---|
| Python | `atexit.register(_auto_synth)` | ✅ Yes |
| TypeScript | `process.on('exit', ...)` | ✅ Yes |
| Go < 1.24 | None | ❌ No |
| Go 1.24+ | `runtime.AddExitHook(...)` | ⚠️ Version-specific |

**Attempted Alternatives** (all failed):

1. **Finalizers** (`runtime.SetFinalizer`):
   - Only runs during garbage collection, not program exit
   - Timing is unpredictable
   - Not guaranteed to run at all

2. **Signal Handlers** (`signal.Notify`):
   - Only catches SIGINT/SIGTERM
   - Doesn't catch normal `os.Exit(0)` or end of `main()`
   - Doesn't work for `go run` termination

3. **Background Goroutine**:
   - Can't detect when `main()` completes
   - Would need channels/sync, adding complexity
   - Still requires user code to signal completion

4. **Go 1.24+ `runtime.AddExitHook`**:
   - Only available in Go 1.24+ (released Q1 2025)
   - Can't use as default until Go 1.26+ is mainstream (~2027)
   - Would require build tags and version-specific code

**Solution**: Accept Go's limitation and provide cleanest possible API

**Implementation in Go**:

```go
// Best possible API given Go's constraints
import stigmeragent "github.com/leftbin/stigmer-sdk/go"

func main() {
    defer stigmeragent.Complete()  // ONE line of boilerplate
    
    agent.New(...)  // Rest is clean
}
```

**Why This is Best Possible**:
- ✅ **One line** - minimal overhead (5 words)
- ✅ **Clear intent** - `Complete()` is self-documenting
- ✅ **Go idiom** - `defer` is standard pattern
- ✅ **Works everywhere** - no version constraints
- ✅ **Predictable** - explicit control flow

**Alternative (without defer) would be much worse**:

```go
// ❌ Without Complete() - manual boilerplate (20+ lines)
func main() {
    agent, _ := agent.New(...)
    
    // User would have to:
    manifest := synth.ToManifest(agent)
    data, _ := proto.Marshal(manifest)
    outputDir := os.Getenv("STIGMER_OUT_DIR")
    if outputDir == "" {
        fmt.Println("Dry-run mode...")
        return
    }
    os.MkdirAll(outputDir, 0755)
    os.WriteFile(filepath.Join(outputDir, "manifest.pb"), data, 0644)
}
```

**Documentation Strategy**:
Created comprehensive documentation explaining this:
- `docs/architecture/synthesis-model.md` (200+ lines)
- Explains Go's language constraints vs Python
- Documents all attempted alternatives
- Justifies why `defer` is necessary
- Shows future path (Go 1.24+)

**Benefits**:
- Users understand WHY (not just HOW)
- No confusion about missing features
- Clear path forward (Go 1.24+ support)
- Manages expectations appropriately

**Future: Go 1.24+ Support**:

```go
//go:build go1.24

package stigmeragent

import "runtime"

func init() {
    // Truly automatic - no defer needed!
    runtime.AddExitHook(synth.AutoSynth)
}
```

Then users on Go 1.24+ can skip the `defer` line entirely.

**Prevention**: 
- Document language constraints prominently
- Don't promise features that require language capabilities not present
- Explain trade-offs clearly in architecture docs
- Plan for future language versions when capabilities arrive
- Accept that different languages have different limitations

**Cross-Language Design Lesson**:
When designing multi-language SDKs, recognize that not all languages have the same capabilities. The "ideal" design in one language may be impossible in another. Document these constraints clearly so users understand the reasoning.

---

## Documentation Organization

**Topic Coverage**: Documentation standards, filename conventions, categorization, navigation

### 2026-01-13 - Following Stigmer Documentation Standards

**Problem**: Created documentation with uppercase filenames (`SYNTHESIS.md`) in wrong locations (root of package instead of `docs/`), not following Stigmer monorepo conventions.

**Root Cause**:
- Didn't check Stigmer documentation standards before creating files
- Default to uppercase for "important" docs (common anti-pattern)
- Placed architectural docs in root instead of categorized location
- Temporary files left in `_cursor/` folder (won't be committed)

**Stigmer Documentation Standards**:

1. **All filenames lowercase with hyphens**:
   - ✅ `synthesis-model.md`
   - ❌ `SYNTHESIS.md`
   - ❌ `SynthesisModel.md`
   - ❌ `synthesis_model.md`

2. **Organized by purpose in `docs/` folder**:
   ```
   docs/
   ├── README.md                    # Documentation index
   ├── getting-started/             # Quick starts, configuration
   ├── architecture/                # System design, patterns
   ├── guides/                      # How-to guides, tutorials
   ├── implementation/              # Implementation details, reports
   └── references/                  # Additional references, notes
   ```

3. **Update `docs/README.md` index** when adding new docs

4. **Never use `_cursor/` for permanent documentation** - it's temporary

**Solution**: Reorganize all documentation to follow standards

**Implementation**:

```bash
# ❌ BEFORE: Non-compliant
go/
├── SYNTHESIS.md                            # Uppercase, wrong location
└── _cursor/
    ├── synthesis-api-improvement.md        # Won't be committed
    └── implementation-summary.md           # Won't be committed

# ✅ AFTER: Standards compliant
go/
└── docs/
    ├── README.md                           # Updated index
    ├── architecture/
    │   └── synthesis-model.md              # Lowercase, categorized
    └── implementation/
        └── synthesis-api-improvement.md    # Lowercase, categorized
```

**Categorization Guidelines**:

| Content Type | Category | Example |
|---|---|---|
| **Why it exists** | `architecture/` | `synthesis-model.md` |
| **What was built** | `implementation/` | `synthesis-api-improvement.md` |
| **How to use** | `guides/` | `getting-started.md` |
| **Quick reference** | `references/` | `proto-mapping.md` |

**Documentation Index Pattern**:

```markdown
## Architecture

- [Synthesis Model](./architecture/synthesis-model.md) - Why Go needs defer
- [Multi-Agent Support](./architecture/multi-agent-support.md) - Multiple agents
```

**Benefits**:
- Consistent across all Stigmer projects
- Easy to find documentation (clear categories)
- Professional open-source conventions
- Scales well as docs grow
- Platform-independent (works everywhere)

**Prevention**: 
- **Always check** `@documentation-standards.md` before creating docs
- Use lowercase-with-hyphens for ALL filenames
- Categorize by purpose (architecture, implementation, guides, references)
- Update `docs/README.md` index immediately
- Never commit files from `_cursor/` - they're temporary

**Automation**:
Could create a pre-commit hook to enforce:
```bash
# Check for uppercase in docs/
find docs/ -name '*[A-Z]*' -type f
```

**Cross-Project Reference**: All Stigmer projects follow same standards

---

## Go Module Management

**Topic Coverage**: go.mod, dependency updates, version constraints

### (Entries will be added as work is done)

**Common patterns to document**:
- Adding proto dependencies
- Module path setup
- Version management

---

## Future Topics

As the Go SDK evolves, add sections for:
- Concurrency Patterns
- Context Management
- Interface Design
- Performance Optimization
- Documentation Generation

---

## Meta: Using This Log

**Good Example**:
- Search for error message: "panic: runtime error: invalid memory address"
- Find section: "Proto Converters"
- Apply documented nil check pattern

**Bad Example**:
- Don't search log
- Spend 30 minutes debugging
- Solve problem that was already documented
- Waste time reinventing solution

**Remember**: This log saves time. Check it first!

---

## Cross-Language References

When a pattern has a Python equivalent:

```markdown
**Python SDK equivalent**: See sdk/python/_rules/.../docs/learning-log.md
for Python approach using `.extend()`

**Go approach**: Use `append()` instead:
```go
agentProto.Spec.SkillRefs = append(agentProto.Spec.SkillRefs, skillRefs...)
```
```

---

## Initial State Note

This learning log is empty because the Go SDK implementation has just begun its self-improvement journey. As you:
- Implement features
- Fix bugs
- Discover patterns
- Solve problems

...this log will grow organically with real-world Go-specific learnings. Each entry will help future implementations (including your own) avoid repeating the same issues and discoveries.

**Start adding entries immediately** when you encounter anything worth documenting!
