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

// Container resolves named dependencies and populates fields tagged with
// `inject`. It is declared here rather than imported so that a CLI never has
// to name a type from the dependency injection library it happens to use.
//
// *di.Container from github.com/mirkobrombin/go-foundation/v2/app/di satisfies
// this interface as it is, so it can be passed straight to WithContainer.
//
// Example:
//
//	container, err := builder.Build()
//	cli.Run(app, cli.WithContainer(container))
type Container interface {
	// Get returns the instance registered under name.
	Get(name string) (any, bool)
	// Has reports whether name is registered.
	Has(name string) bool
	// Inject populates the `inject` tagged fields of target.
	Inject(target any) error
}

// Validator checks a command struct after its flags and arguments have been
// bound and before it runs. A non-nil error aborts the run and prints the help
// for the command.
//
// Example, adapting the validator from go-foundation:
//
//	v := validation.New()
//	cli.Run(app, cli.WithValidator(cli.ValidatorFunc(func(target any) error {
//		if errs := v.Validate(target); len(errs) > 0 {
//			return errs
//		}
//		return nil
//	})))
type Validator interface {
	Validate(target any) error
}

// ValidatorFunc adapts a plain function to the Validator interface.
type ValidatorFunc func(target any) error

// Validate calls f.
func (f ValidatorFunc) Validate(target any) error { return f(target) }
