---
name: cobra
description: >-
  Generates, explains, and debugs CLI applications using Cobra (github.com/spf13/cobra) for Go.
  Use when building command-line tools with subcommands, flags, persistent flags, hooks, 
  shell completions, help generation, argument validation, or migrating from other CLI libraries.
  Covers Command struct, AddCommand, flags (local/persistent), PreRun/PostRun hooks, 
  ValidArgs, completions for bash/zsh/fish/powershell, and integration with viper.
allowed-tools: Read, Grep, Bash(go doc *)
---

# Cobra CLI Framework

Cobra is a library for creating powerful modern CLI applications in Go, used by Kubernetes, Hugo, GitHub CLI, and many more.

## Core Concepts

**Commands** represent actions, **Args** are things, **Flags** are modifiers. Follow the pattern: `APPNAME COMMAND ARG --FLAG`.

## Project Structure

```
appName/
  cmd/
    root.go
    subcommand.go
  main.go
```

## Creating Commands

### Root Command

```go
package cmd

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "Brief description",
    Long:  `Detailed description...`,
    Run: func(cmd *cobra.Command, args []string) {
        // Action here
    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

### Subcommands

```go
var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print version",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("v1.0.0")
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}
```

### Error Handling with RunE

```go
var cmd = &cobra.Command{
    Use: "delete [name]",
    RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) == 0 {
            return fmt.Errorf("name required")
        }
        return deleteResource(args[0])
    },
}
```

## Flags

### Persistent Flags (cascade to children)

```go
var verbose bool

func init() {
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
```

### Local Flags (command-specific)

```go
var source string

func init() {
    cmd.Flags().StringVarP(&source, "source", "s", "", "source directory")
}
```

### Required Flags

```go
func init() {
    rootCmd.Flags().StringVarP(&region, "region", "r", "", "AWS region")
    rootCmd.MarkFlagRequired("region")
}
```

### Flag Groups

```go
func init() {
    rootCmd.Flags().StringVarP(&user, "username", "u", "", "Username")
    rootCmd.Flags().StringVarP(&pass, "password", "p", "", "Password")
    rootCmd.MarkFlagsRequiredTogether("username", "password")
    rootCmd.MarkFlagsMutuallyExclusive("json", "yaml")
}
```

### Count/Array Flags

```go
var verbose int
var files []string

func init() {
    rootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "verbosity (-v, -vv, -vvv)")
    rootCmd.Flags().StringArrayVarP(&files, "file", "f", []string{}, "input files")
}
```

## Argument Validation

```go
var cmd = &cobra.Command{
    Use:  "create [name]",
    Args: cobra.ExactArgs(1),
    Run:  func(cmd *cobra.Command, args []string) { ... },
}
```

Built-in validators:
- `NoArgs` - no positional args allowed
- `ArbitraryArgs` - any number (default)
- `MinimumNArgs(int)` - at least N args
- `MaximumNArgs(int)` - at most N args
- `ExactArgs(int)` - exactly N args
- `RangeArgs(min, max)` - between min and max
- `OnlyValidArgs` - only values in ValidArgs

Combined validation:
```go
Args: cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
```

## PreRun/PostRun Hooks

Execution order: `PersistentPreRun` -> `PreRun` -> `Run` -> `PostRun` -> `PersistentPostRun`

```go
var rootCmd = &cobra.Command{
    Use: "myapp",
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        // Runs before any child command
    },
    PreRun: func(cmd *cobra.Command, args []string) {
        // Runs before this command's Run
    },
    Run: func(cmd *cobra.Command, args []string) {
        // Main logic
    },
    PostRun: func(cmd *cobra.Command, args []string) {
        // Runs after Run
    },
    PersistentPostRun: func(cmd *cobra.Command, args []string) {
        // Runs after any child command
    },
}
```

## Shell Completions

### Static Completions

```go
var cmd = &cobra.Command{
    Use:       "get [resource]",
    ValidArgs: []string{"pod", "service", "deployment"},
    Run:       func(cmd *cobra.Command, args []string) { ... },
}
```

### Dynamic Completions

```go
var cmd = &cobra.Command{
    Use: "status [release]",
    ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
        return getReleases(toComplete), cobra.ShellCompDirectiveNoFileComp
    },
    Run: func(cmd *cobra.Command, args []string) { ... },
}
```

### Flag Completions

```go
cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
    return []cobra.Completion{"json", "yaml", "table"}, cobra.ShellCompDirectiveNoFileComp
})
```

### ShellCompDirective Values

- `ShellCompDirectiveDefault` - default behavior
- `ShellCompDirectiveNoSpace` - no space after completion
- `ShellCompDirectiveNoFileComp` - no file completion
- `ShellCompDirectiveFilterFileExt` - filter by file extension
- `ShellCompDirectiveFilterDirs` - directory names only

## Help Customization

```go
cmd.SetHelpCommand(customHelpCmd)
cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) { ... })
cmd.SetHelpTemplate(customTemplate)
cmd.SetUsageFunc(func(cmd *cobra.Command) error { ... })
cmd.SetUsageTemplate(customTemplate)
```

## Version Flag

```go
var rootCmd = &cobra.Command{
    Use:     "myapp",
    Version: "1.0.0",
}
```

## Command Groups

```go
func init() {
    rootCmd.AddGroup(
        &cobra.Group{ID: "management", Title: "Management Commands:"},
        &cobra.Group{ID: "query", Title: "Query Commands:"},
    )
    
    listCmd.GroupID = "query"
    createCmd.GroupID = "management"
    rootCmd.AddCommand(listCmd, createCmd)
}
```

## Plugin Support (kubectl-style)

```go
rootCmd := &cobra.Command{
    Use: "kubectl-myplugin",
    Annotations: map[string]string{
        cobra.CommandDisplayNameAnnotation: "kubectl myplugin",
    },
}
```

## Common Patterns

### Traverse Children Flags

```go
rootCmd := &cobra.Command{
    Use:              "myapp",
    TraverseChildren: true, // Parse parent flags before child command
}
```

### Silence Errors/Usage

```go
cmd.SilenceErrors = true // Don't print error prefix
cmd.SilenceUsage = true  // Don't print usage on error
```

### Hidden Commands

```go
cmd.Hidden = true
```

### Deprecated Commands

```go
cmd.Deprecated = "use 'newcmd' instead"
```

## Examples Index

| File | Description |
|------|-------------|
| [references/_examples/basic_app.go](references/_examples/basic_app.go) | Minimal CLI application |
| [references/_examples/subcommands.go](references/_examples/subcommands.go) | Nested subcommands |
| [references/_examples/flags.go](references/_examples/flags.go) | Flag patterns |
| [references/_examples/hooks.go](references/_examples/hooks.go) | PreRun/PostRun hooks |
| [references/_examples/completions.go](references/_examples/completions.go) | Shell completions |

## Integration with Viper

```go
import "github.com/spf13/viper"

func init() {
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
    rootCmd.PersistentFlags().StringP("author", "a", "", "author name")
    viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
}
```

## Further Reading

- [User Guide](references/user_guide.md) - Complete user guide documentation
- [Shell Completions](references/completions.md) - Detailed completion documentation
- [Active Help](references/active_help.md) - Context-sensitive help messages
