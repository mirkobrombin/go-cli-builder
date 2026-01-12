package cli

// Runner is an interface for commands that can be run.
type Runner interface {
	Run() error
}

// BeforeRunner is an interface for commands that run before the main run.
type BeforeRunner interface {
	Before() error
}

// AfterRunner is an interface for commands that run after the main run.
type AfterRunner interface {
	After() error
}
