# Flag Handling

Flags in V2 are defined using struct tags on your command structs. The library automatically binds command-line flags to your struct fields.

## Supported Types

The following Go types are supported:
- `bool`: Simple toggle flags (e.g., `--verbose`).
- `string`: String values (e.g., `--name="foo"`).
- `int`, `int64`, etc.: Integer values.
- `float64`: Floating point values.
- `time.Duration`: Duration strings (e.g., `10s`, `1h`).
- `[]string`: Repeated flags (e.g., `--item a --item b`).

## defining Flags

Use the `cli` tag to define a flag. The format is `cli:"name,short_name"`.

```go
type Options struct {
    // Long flag --config, short flag -c
    Config string `cli:"config,c" help:"Path to config file"`
    
    // Long flag --dry-run
    DryRun bool `cli:"dry-run" help:"Simulate execution"`
}
```

## Default Values

Use the `default` tag to specify a default value if the flag is not provided.

```go
Port int `cli:"port" default:"8080"`
```

## Required Flags

Use the `required:"true"` tag to mark a flag as mandatory.

```go
Token string `cli:"token" required:"true"`
```