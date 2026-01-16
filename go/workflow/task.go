package workflow

import (
	"fmt"
	"strings"
)

// TaskKind represents the type of workflow task.
type TaskKind string

// Task kinds matching Zigflow DSL task types.
const (
	TaskKindSet          TaskKind = "SET"
	TaskKindHttpCall     TaskKind = "HTTP_CALL"
	TaskKindGrpcCall     TaskKind = "GRPC_CALL"
	TaskKindSwitch       TaskKind = "SWITCH"
	TaskKindFor          TaskKind = "FOR"
	TaskKindFork         TaskKind = "FORK"
	TaskKindTry          TaskKind = "TRY"
	TaskKindListen       TaskKind = "LISTEN"
	TaskKindWait         TaskKind = "WAIT"
	TaskKindCallActivity TaskKind = "CALL_ACTIVITY"
	TaskKindRaise        TaskKind = "RAISE"
	TaskKindRun          TaskKind = "RUN"
)

// Special task flow control constants.
const (
	// EndFlow indicates the workflow should terminate after this task.
	// Use task.End() method instead of task.Then(EndFlow) for better readability.
	EndFlow = "end"
)

// Task represents a single task in a workflow.
type Task struct {
	// Task name/identifier (must be unique within workflow)
	Name string

	// Task type (determines how to interpret Config)
	Kind TaskKind

	// Task-specific configuration (type depends on Kind)
	Config TaskConfig

	// Export configuration (how to save task output to context)
	ExportAs string

	// Flow control (which task executes next)
	ThenTask string
}

// TaskConfig is a marker interface for task configurations.
type TaskConfig interface {
	isTaskConfig()
}

// Export sets the export directive for this task using a low-level expression.
// For most use cases, prefer ExportAll() or ExportField() for better UX.
// Example: task.Export("${.}") exports entire output.
func (t *Task) Export(expr string) *Task {
	t.ExportAs = expr
	return t
}

// ExportAll exports the entire task output to the workflow context.
// This is a high-level helper that replaces Export("${.}").
// Example: HttpCallTask("fetch",...).ExportAll()
func (t *Task) ExportAll() *Task {
	t.ExportAs = "${.}"
	return t
}

// ExportField exports a specific field from the task output to the workflow context.
// This is a high-level helper that replaces Export("${.field}").
// Example: HttpCallTask("fetch",...).ExportField("count")
func (t *Task) ExportField(fieldName string) *Task {
	t.ExportAs = fmt.Sprintf("${.%s}", fieldName)
	return t
}

// ExportFields exports multiple fields from the task output to the workflow context.
// Each field is exported with its original name.
// Example: HttpCallTask("fetch",...).ExportFields("count", "status", "data")
func (t *Task) ExportFields(fieldNames ...string) *Task {
	// For multiple fields, we export the whole object and let the next task
	// access specific fields. This is more efficient than creating separate exports.
	// In the future, we could support selective field export if the proto supports it.
	t.ExportAs = "${.}"
	return t
}

// Then sets the flow control directive for this task using a task name string.
// Example: task.Then("nextTask") jumps to task named "nextTask".
//
// For type-safe task references, use ThenTask() instead.
func (t *Task) Then(taskName string) *Task {
	t.ThenTask = taskName
	return t
}

// ThenRef sets the flow control directive using a task reference.
// This is type-safe and prevents typos in task names.
//
// Example:
//
//	task1 := workflow.SetTask("init", workflow.SetInt("x", 1))
//	task2 := workflow.HttpCallTask("fetch", ...).ThenRef(task1)
func (t *Task) ThenRef(task *Task) *Task {
	t.ThenTask = task.Name
	return t
}

// End terminates the workflow after this task.
// This is equivalent to task.Then(workflow.EndFlow) but more explicit.
func (t *Task) End() *Task {
	t.ThenTask = EndFlow
	return t
}

// ============================================================================
// SET Task
// ============================================================================

// SetTaskConfig defines the configuration for SET tasks.
type SetTaskConfig struct {
	// Variables to set in workflow state.
	// Keys are variable names, values can be literals or expressions.
	Variables map[string]string
}

func (*SetTaskConfig) isTaskConfig() {}

// SetTask creates a new SET task.
//
// SET tasks assign variables in workflow state.
//
// Example:
//
//	task := workflow.SetTask("init",
//	    workflow.SetVar("apiURL", "https://api.example.com"),
//	    workflow.SetVar("count", "0"),
//	)
func SetTask(name string, opts ...SetTaskOption) *Task {
	cfg := &SetTaskConfig{
		Variables: make(map[string]string),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Task{
		Name:   name,
		Kind:   TaskKindSet,
		Config: cfg,
	}
}

// SetTaskOption is a functional option for configuring SET tasks.
type SetTaskOption func(*SetTaskConfig)

// SetVar adds a variable to a SET task.
// For better type safety, consider using SetInt, SetString, SetBool instead.
func SetVar(key, value string) SetTaskOption {
	return func(cfg *SetTaskConfig) {
		cfg.Variables[key] = value
	}
}

// SetVars adds multiple variables to a SET task.
func SetVars(vars map[string]string) SetTaskOption {
	return func(cfg *SetTaskConfig) {
		for k, v := range vars {
			cfg.Variables[k] = v
		}
	}
}

// SetInt adds an integer variable to a SET task with automatic type conversion.
// This is a high-level helper that provides better UX than SetVar("count", "0").
// Example: SetTask("init", SetInt("count", 0))
func SetInt(key string, value int) SetTaskOption {
	return func(cfg *SetTaskConfig) {
		cfg.Variables[key] = fmt.Sprintf("%d", value)
	}
}

// SetString adds a string variable to a SET task.
// This is semantically clearer than SetVar for string values.
// Example: SetTask("init", SetString("status", "pending"))
func SetString(key, value string) SetTaskOption {
	return func(cfg *SetTaskConfig) {
		cfg.Variables[key] = value
	}
}

// SetBool adds a boolean variable to a SET task with automatic type conversion.
// Example: SetTask("init", SetBool("enabled", true))
func SetBool(key string, value bool) SetTaskOption {
	return func(cfg *SetTaskConfig) {
		cfg.Variables[key] = fmt.Sprintf("%t", value)
	}
}

// SetFloat adds a float variable to a SET task with automatic type conversion.
// Example: SetTask("init", SetFloat("price", 99.99))
func SetFloat(key string, value float64) SetTaskOption {
	return func(cfg *SetTaskConfig) {
		cfg.Variables[key] = fmt.Sprintf("%f", value)
	}
}

// ============================================================================
// HTTP_CALL Task
// ============================================================================

// HttpCallTaskConfig defines the configuration for HTTP_CALL tasks.
type HttpCallTaskConfig struct {
	Method         string            // HTTP method (GET, POST, PUT, DELETE, PATCH)
	URI            string            // HTTP endpoint URI
	Headers        map[string]string // HTTP headers
	Body           map[string]any    // Request body (JSON)
	TimeoutSeconds int32             // Request timeout in seconds
}

func (*HttpCallTaskConfig) isTaskConfig() {}

// HttpCallTask creates a new HTTP_CALL task.
//
// HTTP_CALL tasks make HTTP requests.
//
// Example:
//
//	task := workflow.HttpCallTask("fetchData",
//	    workflow.WithHTTPGet(),  // Type-safe HTTP method
//	    workflow.WithURI("https://api.example.com/data"),
//	    workflow.WithHeader("Authorization", "Bearer ${TOKEN}"),
//	)
func HttpCallTask(name string, opts ...HttpCallTaskOption) *Task {
	cfg := &HttpCallTaskConfig{
		Headers:        make(map[string]string),
		Body:           make(map[string]any),
		TimeoutSeconds: 30, // default timeout
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Task{
		Name:   name,
		Kind:   TaskKindHttpCall,
		Config: cfg,
	}
}

// HttpCallTaskOption is a functional option for configuring HTTP_CALL tasks.
type HttpCallTaskOption func(*HttpCallTaskConfig)

// WithMethod sets the HTTP method using a string.
// For common HTTP methods, prefer using the type-safe helpers:
// WithHTTPGet(), WithHTTPPost(), WithHTTPPut(), WithHTTPPatch(), WithHTTPDelete(), etc.
// Use this function for custom or non-standard HTTP methods.
func WithMethod(method string) HttpCallTaskOption {
	return func(cfg *HttpCallTaskConfig) {
		cfg.Method = method
	}
}

// WithHTTPGet sets the HTTP method to GET.
// This is a type-safe helper for the most common HTTP method.
func WithHTTPGet() HttpCallTaskOption {
	return func(cfg *HttpCallTaskConfig) {
		cfg.Method = "GET"
	}
}

// WithHTTPPost sets the HTTP method to POST.
// This is a type-safe helper for creating or submitting data.
func WithHTTPPost() HttpCallTaskOption {
	return func(cfg *HttpCallTaskConfig) {
		cfg.Method = "POST"
	}
}

// WithHTTPPut sets the HTTP method to PUT.
// This is a type-safe helper for updating or replacing resources.
func WithHTTPPut() HttpCallTaskOption {
	return func(cfg *HttpCallTaskConfig) {
		cfg.Method = "PUT"
	}
}

// WithHTTPPatch sets the HTTP method to PATCH.
// This is a type-safe helper for partial updates to resources.
func WithHTTPPatch() HttpCallTaskOption {
	return func(cfg *HttpCallTaskConfig) {
		cfg.Method = "PATCH"
	}
}

// WithHTTPDelete sets the HTTP method to DELETE.
// This is a type-safe helper for removing resources.
func WithHTTPDelete() HttpCallTaskOption {
	return func(cfg *HttpCallTaskConfig) {
		cfg.Method = "DELETE"
	}
}

// WithHTTPHead sets the HTTP method to HEAD.
// This is a type-safe helper for retrieving headers without body.
func WithHTTPHead() HttpCallTaskOption {
	return func(cfg *HttpCallTaskConfig) {
		cfg.Method = "HEAD"
	}
}

// WithHTTPOptions sets the HTTP method to OPTIONS.
// This is a type-safe helper for retrieving allowed methods and CORS.
func WithHTTPOptions() HttpCallTaskOption {
	return func(cfg *HttpCallTaskConfig) {
		cfg.Method = "OPTIONS"
	}
}

// WithURI sets the HTTP URI.
func WithURI(uri string) HttpCallTaskOption {
	return func(cfg *HttpCallTaskConfig) {
		cfg.URI = uri
	}
}

// WithHeader adds an HTTP header.
func WithHeader(key, value string) HttpCallTaskOption {
	return func(cfg *HttpCallTaskConfig) {
		cfg.Headers[key] = value
	}
}

// WithHeaders adds multiple HTTP headers.
func WithHeaders(headers map[string]string) HttpCallTaskOption {
	return func(cfg *HttpCallTaskConfig) {
		for k, v := range headers {
			cfg.Headers[k] = v
		}
	}
}

// WithBody sets the request body.
func WithBody(body map[string]any) HttpCallTaskOption {
	return func(cfg *HttpCallTaskConfig) {
		cfg.Body = body
	}
}

// WithTimeout sets the request timeout in seconds.
func WithTimeout(seconds int32) HttpCallTaskOption {
	return func(cfg *HttpCallTaskConfig) {
		cfg.TimeoutSeconds = seconds
	}
}

// ============================================================================
// GRPC_CALL Task
// ============================================================================

// GrpcCallTaskConfig defines the configuration for GRPC_CALL tasks.
type GrpcCallTaskConfig struct {
	Service string         // gRPC service name
	Method  string         // gRPC method name
	Body    map[string]any // Request body (proto message as JSON)
}

func (*GrpcCallTaskConfig) isTaskConfig() {}

// GrpcCallTask creates a new GRPC_CALL task.
//
// Example:
//
//	task := workflow.GrpcCallTask("callService",
//	    workflow.WithService("UserService"),
//	    workflow.WithGrpcMethod("GetUser"),
//	    workflow.WithGrpcBody(map[string]any{"userId": "${.userId}"}),
//	)
func GrpcCallTask(name string, opts ...GrpcCallTaskOption) *Task {
	cfg := &GrpcCallTaskConfig{
		Body: make(map[string]any),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Task{
		Name:   name,
		Kind:   TaskKindGrpcCall,
		Config: cfg,
	}
}

// GrpcCallTaskOption is a functional option for configuring GRPC_CALL tasks.
type GrpcCallTaskOption func(*GrpcCallTaskConfig)

// WithService sets the gRPC service name.
func WithService(service string) GrpcCallTaskOption {
	return func(cfg *GrpcCallTaskConfig) {
		cfg.Service = service
	}
}

// WithGrpcMethod sets the gRPC method name.
func WithGrpcMethod(method string) GrpcCallTaskOption {
	return func(cfg *GrpcCallTaskConfig) {
		cfg.Method = method
	}
}

// WithGrpcBody sets the gRPC request body.
func WithGrpcBody(body map[string]any) GrpcCallTaskOption {
	return func(cfg *GrpcCallTaskConfig) {
		cfg.Body = body
	}
}

// ============================================================================
// SWITCH Task
// ============================================================================

// SwitchTaskConfig defines the configuration for SWITCH tasks.
type SwitchTaskConfig struct {
	Cases       []SwitchCase // Conditional cases
	DefaultTask string       // Default task if no cases match
}

// SwitchCase represents a conditional case in a SWITCH task.
type SwitchCase struct {
	Condition string // Condition expression
	Then      string // Task to execute if condition is true
}

func (*SwitchTaskConfig) isTaskConfig() {}

// SwitchTask creates a new SWITCH task.
//
// Example:
//
//	task := workflow.SwitchTask("checkStatus",
//	    workflow.WithCase("${.status == 200}", "processSuccess"),
//	    workflow.WithCase("${.status == 404}", "handleNotFound"),
//	    workflow.WithDefault("handleError"),
//	)
func SwitchTask(name string, opts ...SwitchTaskOption) *Task {
	cfg := &SwitchTaskConfig{
		Cases: []SwitchCase{},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Task{
		Name:   name,
		Kind:   TaskKindSwitch,
		Config: cfg,
	}
}

// SwitchTaskOption is a functional option for configuring SWITCH tasks.
type SwitchTaskOption func(*SwitchTaskConfig)

// WithCase adds a conditional case using a task name string.
// For type-safe task references, use WithCaseRef instead.
// Example: WithCase("${.status == 200}", "handleSuccess")
func WithCase(condition, then string) SwitchTaskOption {
	return func(cfg *SwitchTaskConfig) {
		cfg.Cases = append(cfg.Cases, SwitchCase{
			Condition: condition,
			Then:      then,
		})
	}
}

// WithCaseRef adds a conditional case using a task reference.
// This is type-safe and prevents typos in task names.
// Example: WithCaseRef("${.status == 200}", successTask)
func WithCaseRef(condition string, task *Task) SwitchTaskOption {
	return func(cfg *SwitchTaskConfig) {
		cfg.Cases = append(cfg.Cases, SwitchCase{
			Condition: condition,
			Then:      task.Name,
		})
	}
}

// WithDefault sets the default task using a task name string.
// For type-safe task references, use WithDefaultRef instead.
// Example: WithDefault("handleError")
func WithDefault(task string) SwitchTaskOption {
	return func(cfg *SwitchTaskConfig) {
		cfg.DefaultTask = task
	}
}

// WithDefaultRef sets the default task using a task reference.
// This is type-safe and prevents typos in task names.
// Example: WithDefaultRef(errorTask)
func WithDefaultRef(task *Task) SwitchTaskOption {
	return func(cfg *SwitchTaskConfig) {
		cfg.DefaultTask = task.Name
	}
}

// ============================================================================
// FOR Task
// ============================================================================

// ForTaskConfig defines the configuration for FOR tasks.
type ForTaskConfig struct {
	In string  // Collection expression to iterate over
	Do []Task  // Tasks to execute for each item
}

func (*ForTaskConfig) isTaskConfig() {}

// ForTask creates a new FOR task.
//
// Example:
//
//	task := workflow.ForTask("processItems",
//	    workflow.WithIn("${.items}"),
//	    workflow.WithDo(
//	        workflow.SetTask("process", workflow.SetVar("item", "${.}")),
//	    ),
//	)
func ForTask(name string, opts ...ForTaskOption) *Task {
	cfg := &ForTaskConfig{
		Do: []Task{},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Task{
		Name:   name,
		Kind:   TaskKindFor,
		Config: cfg,
	}
}

// ForTaskOption is a functional option for configuring FOR tasks.
type ForTaskOption func(*ForTaskConfig)

// WithIn sets the collection expression.
func WithIn(expr string) ForTaskOption {
	return func(cfg *ForTaskConfig) {
		cfg.In = expr
	}
}

// WithDo adds tasks to execute for each item.
func WithDo(tasks ...*Task) ForTaskOption {
	return func(cfg *ForTaskConfig) {
		for _, t := range tasks {
			cfg.Do = append(cfg.Do, *t)
		}
	}
}

// ============================================================================
// FORK Task
// ============================================================================

// ForkTaskConfig defines the configuration for FORK tasks.
type ForkTaskConfig struct {
	Branches []ForkBranch // Parallel branches to execute
}

// ForkBranch represents a parallel branch in a FORK task.
type ForkBranch struct {
	Name  string // Branch name
	Tasks []Task // Tasks to execute in this branch
}

func (*ForkTaskConfig) isTaskConfig() {}

// ForkTask creates a new FORK task.
//
// Example:
//
//	task := workflow.ForkTask("parallelProcessing",
//	    workflow.WithBranch("branch1",
//	        workflow.SetTask("task1", workflow.SetVar("x", "1")),
//	    ),
//	    workflow.WithBranch("branch2",
//	        workflow.SetTask("task2", workflow.SetVar("y", "2")),
//	    ),
//	)
func ForkTask(name string, opts ...ForkTaskOption) *Task {
	cfg := &ForkTaskConfig{
		Branches: []ForkBranch{},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Task{
		Name:   name,
		Kind:   TaskKindFork,
		Config: cfg,
	}
}

// ForkTaskOption is a functional option for configuring FORK tasks.
type ForkTaskOption func(*ForkTaskConfig)

// WithBranch adds a parallel branch.
func WithBranch(name string, tasks ...*Task) ForkTaskOption {
	return func(cfg *ForkTaskConfig) {
		branch := ForkBranch{
			Name:  name,
			Tasks: []Task{},
		}
		for _, t := range tasks {
			branch.Tasks = append(branch.Tasks, *t)
		}
		cfg.Branches = append(cfg.Branches, branch)
	}
}

// ============================================================================
// TRY Task
// ============================================================================

// TryTaskConfig defines the configuration for TRY tasks.
type TryTaskConfig struct {
	Tasks []Task       // Tasks to try
	Catch []CatchBlock // Error handlers
}

// CatchBlock represents an error handler in a TRY task.
type CatchBlock struct {
	Errors []string // Error types to catch
	As     string   // Variable name to bind error to
	Tasks  []Task   // Tasks to execute on error
}

func (*TryTaskConfig) isTaskConfig() {}

// TryTask creates a new TRY task.
//
// Example:
//
//	task := workflow.TryTask("handleErrors",
//	    workflow.WithTry(
//	        workflow.HttpCallTask("risky", workflow.WithHTTPGet(), workflow.WithURI("${.url}")),
//	    ),
//	    workflow.WithCatch([]string{"NetworkError"}, "err",
//	        workflow.SetTask("logError", workflow.SetVar("error", "${err}")),
//	    ),
//	)
func TryTask(name string, opts ...TryTaskOption) *Task {
	cfg := &TryTaskConfig{
		Tasks: []Task{},
		Catch: []CatchBlock{},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Task{
		Name:   name,
		Kind:   TaskKindTry,
		Config: cfg,
	}
}

// TryTaskOption is a functional option for configuring TRY tasks.
type TryTaskOption func(*TryTaskConfig)

// WithTry adds tasks to try.
func WithTry(tasks ...*Task) TryTaskOption {
	return func(cfg *TryTaskConfig) {
		for _, t := range tasks {
			cfg.Tasks = append(cfg.Tasks, *t)
		}
	}
}

// WithCatch adds an error handler.
func WithCatch(errors []string, as string, tasks ...*Task) TryTaskOption {
	return func(cfg *TryTaskConfig) {
		catchBlock := CatchBlock{
			Errors: errors,
			As:     as,
			Tasks:  []Task{},
		}
		for _, t := range tasks {
			catchBlock.Tasks = append(catchBlock.Tasks, *t)
		}
		cfg.Catch = append(cfg.Catch, catchBlock)
	}
}

// ============================================================================
// LISTEN Task
// ============================================================================

// ListenTaskConfig defines the configuration for LISTEN tasks.
type ListenTaskConfig struct {
	Event string // Event name to listen for
}

func (*ListenTaskConfig) isTaskConfig() {}

// ListenTask creates a new LISTEN task.
//
// Example:
//
//	task := workflow.ListenTask("waitForApproval",
//	    workflow.WithEvent("approval.granted"),
//	)
func ListenTask(name string, opts ...ListenTaskOption) *Task {
	cfg := &ListenTaskConfig{}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Task{
		Name:   name,
		Kind:   TaskKindListen,
		Config: cfg,
	}
}

// ListenTaskOption is a functional option for configuring LISTEN tasks.
type ListenTaskOption func(*ListenTaskConfig)

// WithEvent sets the event to listen for.
func WithEvent(event string) ListenTaskOption {
	return func(cfg *ListenTaskConfig) {
		cfg.Event = event
	}
}

// ============================================================================
// WAIT Task
// ============================================================================

// WaitTaskConfig defines the configuration for WAIT tasks.
type WaitTaskConfig struct {
	Duration string // Duration to wait (e.g., "5s", "1m", "1h")
}

func (*WaitTaskConfig) isTaskConfig() {}

// WaitTask creates a new WAIT task.
//
// Example:
//
//	task := workflow.WaitTask("delay",
//	    workflow.WithDuration("5s"),
//	)
func WaitTask(name string, opts ...WaitTaskOption) *Task {
	cfg := &WaitTaskConfig{}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Task{
		Name:   name,
		Kind:   TaskKindWait,
		Config: cfg,
	}
}

// WaitTaskOption is a functional option for configuring WAIT tasks.
type WaitTaskOption func(*WaitTaskConfig)

// WithDuration sets the wait duration.
// Accepts both string format and type-safe duration helpers.
//
// String format examples: "5s", "1m", "1h", "1d"
//
// For better type safety and IDE support, consider using duration helpers:
//
//	workflow.WithDuration(workflow.Seconds(5))   // Type-safe
//	workflow.WithDuration(workflow.Minutes(30))  // Discoverable
//	workflow.WithDuration("5s")                  // Also valid
func WithDuration(duration string) WaitTaskOption {
	return func(cfg *WaitTaskConfig) {
		cfg.Duration = duration
	}
}

// ============================================================================
// Duration Builders - Type-safe helpers for time durations
// ============================================================================

// Seconds creates a duration string for the specified number of seconds.
// This is a type-safe helper that replaces manual "Xs" string formatting.
//
// Example:
//
//	workflow.WaitTask("delay",
//	    workflow.WithDuration(workflow.Seconds(5)),  // ✅ Type-safe!
//	)
//
// This replaces the old string-based syntax:
//
//	WithDuration("5s")  // ❌ Not type-safe, typo-prone
//	WithDuration(Seconds(5))  // ✅ Type-safe, discoverable
func Seconds(count int) string {
	return fmt.Sprintf("%ds", count)
}

// Minutes creates a duration string for the specified number of minutes.
// This is a type-safe helper that replaces manual "Xm" string formatting.
//
// Example:
//
//	workflow.WaitTask("delay",
//	    workflow.WithDuration(workflow.Minutes(5)),
//	)
//
// Common use cases:
//   - Polling intervals
//   - Retry delays
//   - Timeout configurations
//   - Rate limiting windows
func Minutes(count int) string {
	return fmt.Sprintf("%dm", count)
}

// Hours creates a duration string for the specified number of hours.
// This is a type-safe helper that replaces manual "Xh" string formatting.
//
// Example:
//
//	workflow.WaitTask("delay",
//	    workflow.WithDuration(workflow.Hours(2)),
//	)
//
// Common use cases:
//   - Long-running batch jobs
//   - Scheduled task delays
//   - Cache expiration
//   - Token validity periods
func Hours(count int) string {
	return fmt.Sprintf("%dh", count)
}

// Days creates a duration string for the specified number of days.
// This is a type-safe helper that replaces manual "Xd" string formatting.
//
// Example:
//
//	workflow.WaitTask("delay",
//	    workflow.WithDuration(workflow.Days(7)),
//	)
//
// Common use cases:
//   - Weekly scheduled tasks
//   - Retention periods
//   - Subscription renewals
//   - Long-term delays
func Days(count int) string {
	return fmt.Sprintf("%dd", count)
}

// ============================================================================
// CALL_ACTIVITY Task
// ============================================================================

// CallActivityTaskConfig defines the configuration for CALL_ACTIVITY tasks.
type CallActivityTaskConfig struct {
	Activity string         // Activity name
	Input    map[string]any // Activity input
}

func (*CallActivityTaskConfig) isTaskConfig() {}

// CallActivityTask creates a new CALL_ACTIVITY task.
//
// Example:
//
//	task := workflow.CallActivityTask("processData",
//	    workflow.WithActivity("DataProcessor"),
//	    workflow.WithActivityInput(map[string]any{"data": "${.data}"}),
//	)
func CallActivityTask(name string, opts ...CallActivityTaskOption) *Task {
	cfg := &CallActivityTaskConfig{
		Input: make(map[string]any),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Task{
		Name:   name,
		Kind:   TaskKindCallActivity,
		Config: cfg,
	}
}

// CallActivityTaskOption is a functional option for configuring CALL_ACTIVITY tasks.
type CallActivityTaskOption func(*CallActivityTaskConfig)

// WithActivity sets the activity name.
func WithActivity(activity string) CallActivityTaskOption {
	return func(cfg *CallActivityTaskConfig) {
		cfg.Activity = activity
	}
}

// WithActivityInput sets the activity input.
func WithActivityInput(input map[string]any) CallActivityTaskOption {
	return func(cfg *CallActivityTaskConfig) {
		cfg.Input = input
	}
}

// ============================================================================
// RAISE Task
// ============================================================================

// RaiseTaskConfig defines the configuration for RAISE tasks.
type RaiseTaskConfig struct {
	Error   string         // Error type/name
	Message string         // Error message
	Data    map[string]any // Additional error data
}

func (*RaiseTaskConfig) isTaskConfig() {}

// RaiseTask creates a new RAISE task.
//
// Example:
//
//	task := workflow.RaiseTask("throwError",
//	    workflow.WithError("ValidationError"),
//	    workflow.WithErrorMessage("Invalid input data"),
//	)
func RaiseTask(name string, opts ...RaiseTaskOption) *Task {
	cfg := &RaiseTaskConfig{
		Data: make(map[string]any),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Task{
		Name:   name,
		Kind:   TaskKindRaise,
		Config: cfg,
	}
}

// RaiseTaskOption is a functional option for configuring RAISE tasks.
type RaiseTaskOption func(*RaiseTaskConfig)

// WithError sets the error type.
func WithError(errorType string) RaiseTaskOption {
	return func(cfg *RaiseTaskConfig) {
		cfg.Error = errorType
	}
}

// WithErrorMessage sets the error message.
func WithErrorMessage(message string) RaiseTaskOption {
	return func(cfg *RaiseTaskConfig) {
		cfg.Message = message
	}
}

// WithErrorData sets additional error data.
func WithErrorData(data map[string]any) RaiseTaskOption {
	return func(cfg *RaiseTaskConfig) {
		cfg.Data = data
	}
}

// ============================================================================
// RUN Task
// ============================================================================

// RunTaskConfig defines the configuration for RUN tasks.
type RunTaskConfig struct {
	WorkflowName string         // Sub-workflow name
	Input        map[string]any // Sub-workflow input
}

func (*RunTaskConfig) isTaskConfig() {}

// RunTask creates a new RUN task.
//
// Example:
//
//	task := workflow.RunTask("executeSubWorkflow",
//	    workflow.WithWorkflow("data-processor"),
//	    workflow.WithWorkflowInput(map[string]any{"data": "${.data}"}),
//	)
func RunTask(name string, opts ...RunTaskOption) *Task {
	cfg := &RunTaskConfig{
		Input: make(map[string]any),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Task{
		Name:   name,
		Kind:   TaskKindRun,
		Config: cfg,
	}
}

// RunTaskOption is a functional option for configuring RUN tasks.
type RunTaskOption func(*RunTaskConfig)

// WithWorkflow sets the sub-workflow name.
func WithWorkflow(workflow string) RunTaskOption {
	return func(cfg *RunTaskConfig) {
		cfg.WorkflowName = workflow
	}
}

// WithWorkflowInput sets the sub-workflow input.
func WithWorkflowInput(input map[string]any) RunTaskOption {
	return func(cfg *RunTaskConfig) {
		cfg.Input = input
	}
}

// ============================================================================
// Variable Interpolation Helpers
// ============================================================================

// VarRef creates a reference to a workflow variable from context.
// Variables set via SetTask are stored in the workflow context and must be
// referenced with dot notation in Serverless Workflow DSL.
//
// Example: WithURI(Interpolate(VarRef("apiURL"), "/data"))
// Generates: ${ .apiURL + "/data" }
//
// Note: This is for variables set in the workflow (via set: tasks).
// For environment variables, use a different helper (future).
func VarRef(varName string) string {
	return fmt.Sprintf("${.%s}", varName)
}

// FieldRef creates a reference to a field in the current context.
// This is a high-level helper that replaces manual "${.field}" syntax.
// Example: SetVar("count", FieldRef("count")) instead of SetVar("count", "${.count}")
func FieldRef(fieldPath string) string {
	return fmt.Sprintf("${.%s}", fieldPath)
}

// Interpolate combines static text with variable references into a valid expression.
// 
// When mixing expressions (${...}) with static strings, this creates a proper
// Serverless Workflow DSL expression using concatenation syntax.
//
// Examples:
//   - Interpolate(VarRef("apiURL"), "/data") 
//     → ${ apiURL + "/data" } ✅
//   - Interpolate("Bearer ", VarRef("token"))
//     → ${ "Bearer " + token } ✅
//   - Interpolate("https://", VarRef("domain"), "/api/v1")
//     → ${ "https://" + domain + "/api/v1" } ✅
//
// Special cases:
//   - Interpolate(VarRef("url")) → ${url} (single expression, no concatenation)
//   - Interpolate("https://api.example.com") → https://api.example.com (plain string)
func Interpolate(parts ...string) string {
	if len(parts) == 0 {
		return ""
	}
	
	// Single part - return as-is
	if len(parts) == 1 {
		return parts[0]
	}
	
	// Check if any part contains an expression (starts with ${)
	hasExpression := false
	for _, part := range parts {
		if strings.HasPrefix(part, "${") {
			hasExpression = true
			break
		}
	}
	
	// If no expressions, just concatenate as plain string
	if !hasExpression {
		return strings.Join(parts, "")
	}
	
	// Build expression with proper concatenation
	exprParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.HasPrefix(part, "${") && strings.HasSuffix(part, "}") {
			// Extract expression content (remove ${ and })
			expr := part[2 : len(part)-1]
			exprParts = append(exprParts, expr)
		} else {
			// Quote static strings
			exprParts = append(exprParts, fmt.Sprintf("\"%s\"", part))
		}
	}
	
	// Join with + operator and wrap in ${ }
	return fmt.Sprintf("${ %s }", strings.Join(exprParts, " + "))
}

// ============================================================================
// Error Field Accessors - Type-safe helpers for accessing caught error fields
// ============================================================================

// ErrorMessage returns a reference to the message field of a caught error.
// This is a type-safe helper that replaces manual "${errorVar.message}" syntax.
//
// When an error is caught in a CATCH block with `as: "errorVar"`, the error object
// contains several fields. ErrorMessage() provides a discoverable way to access
// the human-readable error description.
//
// Example:
//
//	workflow.WithCatchTyped(
//	    workflow.CatchHTTPErrors(),
//	    "httpErr",
//	    workflow.SetTask("handleError",
//	        workflow.SetVar("errorMessage", workflow.ErrorMessage("httpErr")),
//	    ),
//	)
//
// This replaces the old string-based syntax:
//
//	SetVar("errorMessage", "${httpErr.message}")  // ❌ Old way - not discoverable
//	SetVar("errorMessage", ErrorMessage("httpErr")) // ✅ New way - type-safe
func ErrorMessage(errorVar string) string {
	return fmt.Sprintf("${%s.message}", errorVar)
}

// ErrorCode returns a reference to the code field of a caught error.
// This is a type-safe helper that replaces manual "${errorVar.code}" syntax.
//
// The error code is a machine-readable string that indicates the error type.
// This is useful for logging or conditional logic based on error types.
//
// Example:
//
//	workflow.WithCatchTyped(
//	    workflow.CatchAny(),
//	    "err",
//	    workflow.SetTask("logError",
//	        workflow.SetVar("errorCode", workflow.ErrorCode("err")),
//	    ),
//	)
//
// This replaces the old string-based syntax:
//
//	SetVar("errorCode", "${err.code}")  // ❌ Old way - not discoverable
//	SetVar("errorCode", ErrorCode("err")) // ✅ New way - type-safe
func ErrorCode(errorVar string) string {
	return fmt.Sprintf("${%s.code}", errorVar)
}

// ErrorStackTrace returns a reference to the stackTrace field of a caught error.
// This is a type-safe helper that replaces manual "${errorVar.stackTrace}" syntax.
//
// The stack trace provides debugging information about where the error occurred.
// This is optional and may not be present for all error types.
//
// Example:
//
//	workflow.WithCatchTyped(
//	    workflow.CatchAny(),
//	    "err",
//	    workflow.SetTask("logError",
//	        workflow.SetVar("errorStackTrace", workflow.ErrorStackTrace("err")),
//	    ),
//	)
//
// This replaces the old string-based syntax:
//
//	SetVar("errorStackTrace", "${err.stackTrace}")  // ❌ Old way - not discoverable
//	SetVar("errorStackTrace", ErrorStackTrace("err")) // ✅ New way - type-safe
func ErrorStackTrace(errorVar string) string {
	return fmt.Sprintf("${%s.stackTrace}", errorVar)
}

// ErrorObject returns a reference to the entire caught error object.
// This is a type-safe helper that replaces manual "${errorVar}" syntax.
//
// Use this when you want to pass the entire error object (with all fields)
// to another task, such as logging or external error tracking services.
//
// Example:
//
//	workflow.WithCatchTyped(
//	    workflow.CatchAny(),
//	    "err",
//	    workflow.HttpCallTask("reportError",
//	        workflow.WithHTTPPost(),
//	        workflow.WithURI("https://api.example.com/errors"),
//	        workflow.WithBody(map[string]any{
//	            "error": workflow.ErrorObject("err"), // Pass entire error
//	            "workflow": "data-pipeline",
//	        }),
//	    ),
//	)
//
// This replaces the old string-based syntax:
//
//	"error": "${err}"  // ❌ Old way - not discoverable
//	"error": ErrorObject("err") // ✅ New way - type-safe
func ErrorObject(errorVar string) string {
	return fmt.Sprintf("${%s}", errorVar)
}

// ============================================================================
// Arithmetic Expression Builders - Common patterns for computed values
// ============================================================================

// Increment returns an expression that adds 1 to a variable.
// This is a high-level helper for the extremely common pattern of incrementing counters.
//
// Use this for retry counters, iteration counts, and other increment scenarios.
//
// Example:
//
//	workflow.SetTask("retry",
//	    workflow.SetVar("retryCount", workflow.Increment("retryCount")),
//	)
//
// This replaces the old string-based syntax:
//
//	SetVar("retryCount", "${retryCount + 1}")  // ❌ Old way - not discoverable
//	SetVar("retryCount", Increment("retryCount")) // ✅ New way - type-safe
//
// Common use cases:
//   - Retry counters in error handling
//   - Loop iteration counters
//   - Attempt tracking
//   - Step numbering
func Increment(varName string) string {
	return fmt.Sprintf("${%s + 1}", varName)
}

// Decrement returns an expression that subtracts 1 from a variable.
// This is a high-level helper for the common pattern of decrementing counters.
//
// Use this for countdown timers, remaining items, and other decrement scenarios.
//
// Example:
//
//	workflow.SetTask("processItem",
//	    workflow.SetVar("remaining", workflow.Decrement("remaining")),
//	)
//
// This replaces the old string-based syntax:
//
//	SetVar("remaining", "${remaining - 1}")  // ❌ Old way - not discoverable
//	SetVar("remaining", Decrement("remaining")) // ✅ New way - type-safe
//
// Common use cases:
//   - Countdown timers
//   - Remaining items tracking
//   - Capacity tracking
//   - Quota management
func Decrement(varName string) string {
	return fmt.Sprintf("${%s - 1}", varName)
}

// Expr provides an escape hatch for complex expressions that don't have dedicated helpers.
// Use this when you need arithmetic, string concatenation, or other computations
// that aren't covered by simple helpers like Increment() or Decrement().
//
// This is the "progressive disclosure" pattern - simple things use helpers,
// complex things use expressions directly.
//
// Example:
//
//	// Complex arithmetic
//	workflow.SetVar("total", workflow.Expr("(price * quantity) + tax"))
//
//	// String concatenation
//	workflow.SetVar("fullName", workflow.Expr("firstName + ' ' + lastName"))
//
//	// Conditional expressions
//	workflow.SetVar("status", workflow.Expr("score >= 90 ? 'A' : 'B'"))
//
// Note: For simple cases, prefer dedicated helpers:
//   - Use Increment("x") instead of Expr("x + 1")
//   - Use VarRef("name") instead of Expr("name") for simple references
//   - Use ErrorMessage("err") instead of Expr("err.message") for error fields
func Expr(expression string) string {
	return fmt.Sprintf("${%s}", expression)
}

// ============================================================================
// Condition Builders - High-level helpers for building conditional expressions
// ============================================================================

// Field returns a field reference expression (without ${} wrapper) for use in conditions.
// This is specifically for condition builders. For variable interpolation, use FieldRef().
// Example: Field("status") returns ".status"
func Field(fieldPath string) string {
	return fmt.Sprintf(".%s", fieldPath)
}

// Var returns a variable reference expression (without ${} wrapper) for use in conditions.
// This is specifically for condition builders. For variable interpolation, use VarRef().
// Example: Var("apiURL") returns "apiURL"
func Var(varName string) string {
	return varName
}

// Literal returns a literal value wrapped in quotes for use in conditions.
// Example: Literal("200") returns "\"200\""
func Literal(value string) string {
	return fmt.Sprintf("\"%s\"", value)
}

// Number returns a numeric literal for use in conditions (no quotes).
// Example: Number(200) returns "200"
func Number(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

// Equals builds an equality condition expression.
// Example: Equals(Field("status"), Number(200)) generates "${.status == 200}"
func Equals(left, right string) string {
	return fmt.Sprintf("${%s == %s}", left, right)
}

// NotEquals builds an inequality condition expression.
// Example: NotEquals(FieldRef("status"), "200") generates "${.status != 200}"
func NotEquals(left, right string) string {
	return fmt.Sprintf("${%s != %s}", left, right)
}

// GreaterThan builds a greater-than condition expression.
// Example: GreaterThan(FieldRef("count"), "10") generates "${.count > 10}"
func GreaterThan(left, right string) string {
	return fmt.Sprintf("${%s > %s}", left, right)
}

// GreaterThanOrEqual builds a greater-than-or-equal condition expression.
// Example: GreaterThanOrEqual(FieldRef("status"), "500") generates "${.status >= 500}"
func GreaterThanOrEqual(left, right string) string {
	return fmt.Sprintf("${%s >= %s}", left, right)
}

// LessThan builds a less-than condition expression.
// Example: LessThan(FieldRef("count"), "100") generates "${.count < 100}"
func LessThan(left, right string) string {
	return fmt.Sprintf("${%s < %s}", left, right)
}

// LessThanOrEqual builds a less-than-or-equal condition expression.
// Example: LessThanOrEqual(FieldRef("count"), "100") generates "${.count <= 100}"
func LessThanOrEqual(left, right string) string {
	return fmt.Sprintf("${%s <= %s}", left, right)
}

// And combines multiple conditions with logical AND.
// Example: And(Equals(FieldRef("status"), "200"), Equals(FieldRef("type"), "success"))
func And(conditions ...string) string {
	// Remove ${ and } wrappers from conditions for proper nesting
	unwrapped := make([]string, len(conditions))
	for i, cond := range conditions {
		unwrapped[i] = strings.TrimPrefix(strings.TrimSuffix(cond, "}"), "${")
	}
	return fmt.Sprintf("${%s}", strings.Join(unwrapped, " && "))
}

// Or combines multiple conditions with logical OR.
// Example: Or(Equals(FieldRef("status"), "200"), Equals(FieldRef("status"), "201"))
func Or(conditions ...string) string {
	// Remove ${ and } wrappers from conditions for proper nesting
	unwrapped := make([]string, len(conditions))
	for i, cond := range conditions {
		unwrapped[i] = strings.TrimPrefix(strings.TrimSuffix(cond, "}"), "${")
	}
	return fmt.Sprintf("${%s}", strings.Join(unwrapped, " || "))
}

// Not negates a condition.
// Example: Not(Equals(FieldRef("status"), "200")) generates "${!(.status == 200)}"
func Not(condition string) string {
	// Remove ${ and } wrapper from condition for proper nesting
	unwrapped := strings.TrimPrefix(strings.TrimSuffix(condition, "}"), "${")
	return fmt.Sprintf("${!(%s)}", unwrapped)
}
