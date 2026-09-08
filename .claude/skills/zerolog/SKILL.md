---
name: zerolog
description: Generates, explains, and debugs structured logging code using zerolog (github.com/rs/zerolog) for Go. Use when writing zero-allocation JSON loggers, configuring log levels, adding structured fields, setting up console/file/multi writers, implementing hooks, using context-based logging, sampling high-volume logs, or logging errors with stack traces.
allowed-tools: Read, Grep, Bash(go doc *)
---

# Zerolog - Zero Allocation JSON Logger

Zerolog provides high-performance structured logging with zero allocations for most operations. Use this skill when implementing logging in Go applications that require speed, structured output, or JSON compatibility.

## Quick Start

```go
import (
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func main() {
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
    log.Info().Str("key", "value").Msg("hello world")
}
```

## Log Levels

From highest to lowest: `panic`, `fatal`, `error`, `warn`, `info`, `debug`, `trace`

```go
zerolog.SetGlobalLevel(zerolog.InfoLevel)

if e := log.Debug(); e.Enabled() {
    e.Str("computed", expensiveValue()).Msg("debug info")
}
```

## Structured Fields

### Standard Types
```go
log.Info().
    Str("string", "value").
    Int("count", 42).
    Float64("ratio", 3.14).
    Bool("enabled", true).
    Msg("message")
```

### Advanced Fields
```go
log.Error().
    Err(err).
    Stack().
    Dur("duration", time.Since(start)).
    Time("timestamp", time.Now()).
    Msg("operation failed")

log.Info().
    Dict("user", log.Info().CreateDict().
        Str("name", "John").
        Int("id", 123)).
    Msg("user action")
```

## Logger Configuration

### Basic Logger
```go
logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
```

### Sub-loggers with Context
```go
sublogger := log.With().Str("component", "api").Logger()
sublogger.Info().Msg("request handled")
```

### Custom Field Names
```go
zerolog.TimestampFieldName = "t"
zerolog.LevelFieldName = "l"
zerolog.MessageFieldName = "m"
zerolog.ErrorFieldName = "err"
```

## Output Writers

### ConsoleWriter (Human-Readable)
```go
log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
log.Info().Str("foo", "bar").Msg("Hello World")
```

### MultiLevelWriter
```go
console := zerolog.ConsoleWriter{Out: os.Stdout}
multi := zerolog.MultiLevelWriter(console, os.Stdout)
logger := zerolog.New(multi).With().Timestamp().Logger()
```

### File Output
```go
file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
logger := zerolog.New(file).With().Timestamp().Logger()
```

### Non-Blocking Diode Writer
```go
wr := diode.NewWriter(os.Stdout, 1000, 10*time.Millisecond, func(missed int) {
    fmt.Printf("Dropped %d messages\n", missed)
})
logger := zerolog.New(wr)
```

## Context Integration

### Attach Logger to Context
```go
ctx := logger.WithContext(ctx)
logger := zerolog.Ctx(ctx)
logger.Info().Msg("from context")
```

### Pass Context to Events (for hooks)
```go
type TracingHook struct{}

func (h TracingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
    ctx := e.GetCtx()
    spanID := getSpanID(ctx)
    e.Str("span_id", spanID)
}

logger.Hook(TracingHook{}).Info().Ctx(ctx).Msg("traced")
```

## Error Logging with Stack Traces

```go
import "github.com/rs/zerolog/pkgerrors"

zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

err := errors.New("something failed")
log.Error().Stack().Err(err).Msg("")
```

## Hooks

```go
type SeverityHook struct{}

func (h SeverityHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
    if level != zerolog.NoLevel {
        e.Str("severity", level.String())
    }
}

hooked := log.Hook(SeverityHook{})
hooked.Warn().Msg("alert")
```

## Log Sampling

### Basic Sampling
```go
sampled := log.Sample(&zerolog.BasicSampler{N: 10})
sampled.Info().Msg("logs every 10th message")
```

### Burst Sampling
```go
sampled := log.Sample(zerolog.LevelSampler{
    DebugSampler: &zerolog.BurstSampler{
        Burst: 5,
        Period: 1*time.Second,
        NextSampler: &zerolog.BasicSampler{N: 100},
    },
})
```

### Predefined Samplers
```go
log.Sample(zerolog.Often)    // ~1 in 10
log.Sample(zerolog.Sometimes) // ~1 in 100
log.Sample(zerolog.Rarely)    // ~1 in 1000
```

## HTTP Integration (hlog)

```go
import "github.com/rs/zerolog/hlog"

log := zerolog.New(os.Stdout).With().Timestamp().Logger()

handler := hlog.NewHandler(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    hlog.FromRequest(r).Info().Str("path", r.URL.Path).Msg("request")
}))

handler = hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
    hlog.FromRequest(r).Info().
        Int("status", status).
        Dur("duration", duration).
        Msg("access")
})(handler)
```

## Custom Object Marshaling

```go
type User struct {
    Name string
    ID   int
}

func (u User) MarshalZerologObject(e *zerolog.Event) {
    e.Str("name", u.Name).Int("id", u.ID)
}

log.Info().Object("user", User{Name: "Alice", ID: 1}).Msg("action")
```

## Important Patterns

### Always Call Msg or Send
```go
log.Info().Str("key", "val").Msg("done")
log.Info().Str("key", "val").Send()
```

### Conditional Logging
```go
if e := log.Debug(); e.Enabled() {
    e.Str("expensive", compute()).Msg("debug")
}
```

### Concurrency Safety
```go
func handler(w http.ResponseWriter, r *http.Request) {
    logger := log.Logger.With().Logger()
    logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
        return c.Str("request_id", id)
    })
}
```

## References

See [references/_examples/](references/_examples/) for complete working examples demonstrating all features.
