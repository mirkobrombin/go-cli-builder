# Command Aliases

You can define aliases for your commands using the `aliases` tag (comma-separated).

## Usage

```go
type Root struct {
    // "rm" is an alias for "remove"
    Remove RemoveCmd `cmd:"remove" aliases:"rm,del"`
}
```

This allows users to invoke the command using either `remove`, `rm`, or `del`.