package workflow

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

// Export sets the export directive for this task.
// Example: task.Export("${.}") exports entire output.
func (t *Task) Export(expr string) *Task {
	t.ExportAs = expr
	return t
}

// Then sets the flow control directive for this task.
// Example: task.Then("nextTask") jumps to task named "nextTask".
func (t *Task) Then(taskName string) *Task {
	t.ThenTask = taskName
	return t
}

// End terminates the workflow after this task.
func (t *Task) End() *Task {
	t.ThenTask = "end"
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
//	    workflow.WithMethod("GET"),
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

// WithMethod sets the HTTP method.
func WithMethod(method string) HttpCallTaskOption {
	return func(cfg *HttpCallTaskConfig) {
		cfg.Method = method
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

// WithCase adds a conditional case.
func WithCase(condition, then string) SwitchTaskOption {
	return func(cfg *SwitchTaskConfig) {
		cfg.Cases = append(cfg.Cases, SwitchCase{
			Condition: condition,
			Then:      then,
		})
	}
}

// WithDefault sets the default task.
func WithDefault(task string) SwitchTaskOption {
	return func(cfg *SwitchTaskConfig) {
		cfg.DefaultTask = task
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
//	        workflow.HttpCallTask("risky", workflow.WithMethod("GET"), workflow.WithURI("${.url}")),
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
func WithDuration(duration string) WaitTaskOption {
	return func(cfg *WaitTaskConfig) {
		cfg.Duration = duration
	}
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
