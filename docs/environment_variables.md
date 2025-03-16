# Environment Variable Handling

The Go CLI Builder library provides convenient functions to retrieve values 
from environment variables. This is useful for configuring your CLI application 
without relying solely on command-line flags.

## Retrieving Environment Variable Values

The `command.Command` struct offers the following methods for accessing 
environment variables:

- `GetEnv(key, defaultValue string) string`: Retrieves the value of the environment variable specified by `key`. If the variable is not set, it returns the `defaultValue`.
- `GetEnvInt(key string, defaultValue int) int`: Retrieves the integer value of the environment variable specified by `key`. If the variable is not set or its value is not a valid integer, it returns the `defaultValue`.
- `GetEnvBool(key string, defaultValue bool`: Retrieves the boolean value of the environment variable specified by `key`. If the variable is not set or its value is not a valid boolean, it returns the `defaultValue`.

**Example:**

```go
func runExample(cmd *command.Command, rootFlags *command.RootFlags, args []string) error {
	apiKey := cmd.GetEnv("API_KEY", "default_api_key")
	port := cmd.GetEnvInt("PORT", 8080)
	debugMode := cmd.GetEnvBool("DEBUG", false)

	cmd.Logger.Info("API Key: %s", apiKey)
	cmd.Logger.Info("Port: %d", port)
	cmd.Logger.Info("Debug Mode: %t", debugMode)

	// ...
	return nil
}
```

These functions simplify the process of reading environment variables and 
handling cases where they might not be set or have an invalid format.