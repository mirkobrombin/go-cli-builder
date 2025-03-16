package cli

// Runner is the interface that commands must implement to be executed.
type Runner interface {
	Run() error
}

// BeforeRunner is the interface that commands can implement to run logic before execution.
type BeforeRunner interface {
	Before() error
}

// AfterRunner is the interface that commands can implement to run logic after execution.
type AfterRunner interface {
	After() error
}
