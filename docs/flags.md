# Flag Handling

This document explains how to add and retrieve command-line flags using the 
Go CLI Builder library.

## Adding Flags

The `command.Command` struct provides methods to add different types of flags:

- `AddFlag(name, shortName, usage, defaultValue string, allowArg bool) *string`: Adds a string flag.
    - `name`: The long name of the flag (e.g., "name").
    - `shortName`: The short name of the flag (e.g., "n").
    - `usage`: The description of the flag for the help message.
    - `defaultValue`: The default value of the flag.
    - `allowArg`: A boolean indicating whether the flag expects an argument.
    - Returns a pointer to the string value of the flag.

- `AddIntFlag(name, shortName, usage string, defaultValue int, allowArg bool) *int`: Adds an integer flag.
    - Parameters are similar to `AddFlag`, but for integer values.
    - Returns a pointer to the integer value of the flag.

- `AddBoolFlag(name, shortName, usage string, defaultValue bool, allowArg bool) *bool`: Adds a boolean flag.
    - Parameters are similar to `AddFlag`, but for boolean values.
    - Returns a pointer to the boolean value of the flag.

**Example:**

```go
cmd := &command.Command{
    // ...
}
nameFlag := cmd.AddFlag("name", "n", "The name of the item", "", true)
verboseFlag := cmd.AddBoolFlag("verbose", "v", "Enable verbose output", false, false)
countFlag := cmd.AddIntFlag("count", "c", "The number of items", 1, true)
```

## Retrieving Flag Values

Once the flags are parsed, you can retrieve their values using the following 
methods of the `command.Command` struct:

- `GetFlagString(name string) string`: Returns the string value of the flag. Returns an empty string if the flag is not set.
- `GetFlagInt(name string) int`: Returns the integer value of the flag. Returns -1 if the flag is not set or is not an integer.
- `GetFlagBool(name string) bool`: Returns the boolean value of the flag. Returns false if the flag is not set or is not a boolean.

**Example:**

```go
func runExample(cmd *command.Command, rootFlags *command.RootFlags, args []string) error {
	name := cmd.GetFlagString("name")
	verbose := cmd.GetFlagBool("verbose")
	count := cmd.GetFlagInt("count")
	// ...
	return nil
}
```

## Short Flag Names

The library supports short names for flags. When defining a flag, you can 
provide a short name (a single character). Users can then use either the long 
name (e.g., `--name`) or the short name (e.g., `-n`) to specify the flag.

## Root Flags

Root flags are defined on the `RootCommand` and are parsed before any 
subcommand. Their values can be accessed by subcommands through the 
`rootFlags` parameter in the `Run` function.

The `command.RootFlags` struct provides methods to retrieve the values of 
root flags:

- `GetString(name string) string`
- `GetInt(name string) int`
- `GetBool(name string) bool`

**Example:**

```go
// In main.go
rootCmd := root.NewRootCommand(...)
verboseRootFlag := rootCmd.AddBoolFlag("verbose", "v", "Enable verbose output", false, false)

// In a subcommand's Run function
func runSubcommand(cmd *command.Command, rootFlags *command.RootFlags, args []string) error {
	if rootFlags.GetBool("verbose") {
		cmd.Logger.Info("Verbose output enabled")
	}
	// ...
	return nil
}
```