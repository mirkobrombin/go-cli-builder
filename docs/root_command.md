# Root Command

The Root Command is simply the top-level struct passed to `cli.Run()`. It serves as the entry point for your application and typically holds global flags and subcommands.

## Usage

```go
type Root struct {
    // Global flags
    Debug bool `cli:"debug"`
    
    // Subcommands
    Serve ServeCmd `cmd:"serve"`
}

func main() {
    app := &Root{}
    cli.Run(app)
}
```

Any flags defined on the Root struct are considered **global** and are available (via dependency injection or passing) down the tree, provided you handle the propagation logic or structure your app accordingly. Currently, `cli.Base` provides context, but global flag values are bound to the Root struct instance.