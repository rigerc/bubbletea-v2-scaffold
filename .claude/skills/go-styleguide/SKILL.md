---
name: go-styleguide
description: >-
  Applies Google's Go Style Guide when writing, reviewing, or refactoring Go code.
  Covers naming conventions, formatting, error handling, package design, and testing
  patterns to produce readable, idiomatic, maintainable Go. Use when writing new Go
  code, reviewing Go PRs, or auditing a Go codebase for style issues.
user-invocable: true
allowed-tools: Read, Grep, Bash(go doc *)
metadata:
  source: https://google.github.io/styleguide/go
  version: "1.0"
  category: go
---

# Go Style Guide

Based on the [Google Go Style Guide](https://google.github.io/styleguide/go).
Full reference docs are in [references/](references/).

---

## Core Principles (in priority order)

1. **Clarity** — purpose and rationale are obvious to the reader
2. **Simplicity** — accomplish goals in the simplest way possible
3. **Concision** — high signal-to-noise ratio; no repetitive or extraneous code
4. **Maintainability** — easy to change correctly; APIs that grow gracefully
5. **Consistency** — matches the surrounding codebase; consistency breaks ties

When in doubt, optimize for the reader, not the author.

---

## Formatting

- Run `gofmt` (or `goimports`). All code must be `gofmt`-formatted. No exceptions.
- No fixed line length. If a line feels too long, **refactor** rather than wrap.
- Do not split lines before indentation changes (function declarations, conditionals).
- Do not split long strings (e.g., URLs) across multiple lines.

---

## Naming

### General rules
- Use `MixedCaps` / `mixedCaps` (camel case). No `snake_case`, no `ALL_CAPS`.
- Names should be short in small scopes, longer in large scopes.
- Name based on role/meaning, not type or value.

### Packages
- Lowercase only, no underscores (e.g., `tabwriter`, not `tab_writer`).
- Avoid generic names: `util`, `helper`, `common`, `model`, `handler`.
- Do not repeat the package name in function names: `yamlconfig.Parse`, not `yamlconfig.ParseYAMLConfig`.

### Functions and methods
- Returning something → noun-like name: `Config.JobName()`
- Doing something → verb-like name: `Config.WriteTo()`
- No `Get`/`get` prefix: use `Counts()` not `GetCounts()`. Use `Compute`/`Fetch` when the call is expensive or remote.
- Do not repeat the receiver type in the method name: `WriteTo`, not `WriteConfigTo`.
- Do not repeat parameter names in the function name: `Override(dest, src)` not `OverrideFirstWithSecond`.

### Receivers
- Short (1-2 letters), abbreviation of the type: `func (c *Config)`, `func (ri *ResearchInfo)`.
- Never `this` or `self`.
- Consistent across all methods of a type.

### Constants
- `MixedCaps` like everything else: `MaxPacketSize`, not `MAX_PACKET_SIZE` or `kMaxBufferSize`.
- Name based on role, not value.

### Initialisms / Acronyms
- Keep consistent case within the initialism: `URL`, `ID`, `DB`, `HTTP`, `GRPC`.
- `urlPony` (unexported) or `URLPony` (exported), never `UrlPony`.

### Variables
- Scope-proportional length: single letter ok in tiny scopes, multi-word needed at file scope.
- Avoid shadowing well-known identifiers (`err`, `ctx`, `ok`).

---

## Error Handling

- Always handle errors. Never assign to `_` unless intentional and documented.
- Return errors as the last return value.
- Wrap errors with context: `fmt.Errorf("loading config: %w", err)`.
- Use `%w` (not `%v`) when the caller may need to `errors.Is` / `errors.As`.
- Error strings are lowercase and have no trailing punctuation (they get composed).
- Sentinel errors use `errors.New`: `var ErrNotFound = errors.New("not found")`.
- Custom error types implement the `error` interface via `Error() string`.

```go
// Good:
if err := doSomething(); err != nil {
    return fmt.Errorf("doing something: %w", err)
}

// Good — signal a non-obvious condition:
if err := doSomething(); err == nil { // if NO error
    // ...
}
```

---

## Packages and Imports

- Group imports: stdlib → external → internal. Use `goimports` to manage.
- Rename imports only when necessary to avoid collisions; use the same alias consistently across files.
- Avoid import cycles.
- Avoid "utility" packages that become catch-alls. Prefer focused, well-named packages.

---

## Interfaces

- Define interfaces at the point of use (consumer side), not in the package that implements them.
- Keep interfaces small. Prefer one-method interfaces.
- Do not add interfaces preemptively — only when you have multiple implementations or need to abstract for testing.
- Document what the interface guarantees, not just what it has.

---

## Testing

- Use table-driven tests for parameterized cases.
- Name test functions: `TestFoo`, `TestFoo_condition`, `BenchmarkFoo`.
- Test package naming: `package foo` (whitebox) or `package foo_test` (blackbox).
- Prefer `t.Errorf` over `t.Fatalf` unless the test truly cannot continue.
- Provide clear failure messages: `t.Errorf("Foo(%v) = %v, want %v", input, got, want)`.
- Test helper packages: append `test` to the production package name (e.g., `creditcardtest`).

```go
// Good: table-driven test
tests := []struct {
    name  string
    input string
    want  int
}{
    {"empty", "", 0},
    {"single", "a", 1},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got := Len(tt.input)
        if got != tt.want {
            t.Errorf("Len(%q) = %d, want %d", tt.input, got, tt.want)
        }
    })
}
```

---

## Concurrency

- Prefer channels for communicating between goroutines; prefer mutexes for protecting shared state.
- Document goroutine lifetimes. Make it clear who is responsible for stopping a goroutine.
- Use `context.Context` for cancellation and deadline propagation; pass it as the first parameter.
- Name context variable `ctx` consistently.

---

## Complexity & Abstractions

- Use the simplest mechanism sufficient: core language construct → stdlib → well-known library.
- Add complexity deliberately and document why.
- Do not add abstractions (interfaces, helper functions) until they're needed by multiple callers.
- A helper that hides critical logic makes future bugs more likely.

---

## Common Pitfalls

- Don't use `=` where `:=` is needed (or vice versa) — the difference can be subtle.
- Closures capturing loop variables: capture by copy or use `t.Parallel()` patterns.
- Avoid named return values except in short functions where they aid documentation.
- Don't ignore the second return value from map lookups when the zero value is meaningful.

---

## Reference Documents

| Document | Purpose |
|----------|---------|
| [references/guide.md](references/guide.md) | Core style guide (normative, canonical) |
| [references/decisions.md](references/decisions.md) | Detailed style decisions with rationale |
| [references/best-practices.md](references/best-practices.md) | Patterns for common situations |
| [references/index.md](references/index.md) | Overview and definitions |
