# Environment Variables

You can map flags to environment variables using the `env` tag. If the flag is not provided on the command line, the library will look up the specified environment variable.

## Usage

```go
type AuthCmd struct {
    ApiKey string `cli:"api-key" env:"API_KEY" help:"API Key for authentication"`
}
```

In this example, if the user does not provide `--api-key`, the library will check the `API_KEY` environment variable.

**Priority Order:**
1. Command Line Flag
2. Environment Variable
3. Default Value