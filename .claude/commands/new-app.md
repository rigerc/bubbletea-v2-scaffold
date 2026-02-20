Create a new BubbleTea v2 application from the scaffold.

## Arguments

App name: `$ARGUMENTS`

---

## Step 1 — Validate the app name

The provided name is: **$ARGUMENTS**

- If `$ARGUMENTS` is empty, stop immediately and tell the user:
  `Usage: /new-app <app-name>  (e.g. /new-app myapp)`
- The name must be a valid Go identifier component: lowercase letters, digits,
  and hyphens only, starting with a letter, no spaces or special characters.
- If invalid, stop and explain the problem.

---

## Step 2 — Check the destination

- Source: `scaffold/`  (relative to the repo root)
- Destination: `$ARGUMENTS/`  (a new sibling directory next to `scaffold/`)

If `$ARGUMENTS/` already exists, stop and tell the user.

---

## Step 3 — Copy the scaffold

Copy the entire `scaffold/` directory tree to `$ARGUMENTS/`, preserving all
subdirectories and files. This includes hidden files (`.gitignore`,
`.golangci.yml`, `.goreleaser.yaml`, `.github/`).

---

## Step 4 — Rewrite `go.mod`

In `$ARGUMENTS/go.mod`, change the first line only:

```
module scaffold
```
→
```
module $ARGUMENTS
```

Leave the `go` directive and all `require` / `replace` blocks exactly as-is.

---

## Step 5 — Update internal import paths in all Go files

In **every** `.go` file under `$ARGUMENTS/`, replace every occurrence of the
import path prefix:

```
"scaffold/
```
→
```
"$ARGUMENTS/
```

This covers all internal imports such as:
- `"scaffold/cmd"` → `"$ARGUMENTS/cmd"`
- `"scaffold/config"` → `"$ARGUMENTS/config"`
- `"scaffold/internal/logger"` → `"$ARGUMENTS/internal/logger"`
- `"scaffold/internal/ui"` → `"$ARGUMENTS/internal/ui"`
- etc.

Do **not** change any other strings — only the quoted import path prefix.

---

## Step 6 — Update `cmd/root.go`

Make these targeted changes in `$ARGUMENTS/cmd/root.go`:

1. **`Use` field** — the CLI binary name:
   ```go
   Use:   "scaffold",
   ```
   →
   ```go
   Use:   "$ARGUMENTS",
   ```

2. **Example invocations** — every line inside the `Example:` backtick string
   that starts with `  scaffold` (two-space indent, then the word scaffold):
   replace `scaffold` with `$ARGUMENTS` on those lines only.
   Example: `  scaffold --debug` → `  $ARGUMENTS --debug`

3. **Config file default hint** in `PersistentFlags`:
   ```go
   "Path to configuration file (default: $HOME/.scaffold.json)"
   ```
   →
   ```go
   "Path to configuration file (default: $HOME/.$ARGUMENTS.json)"
   ```

4. **`Long` description opening** — the first word of the Long string is the
   binary name:
   Replace only the very first occurrence of `scaffold` in the Long string
   (i.e. `scaffold is a comprehensive scaffold…` → `$ARGUMENTS is a
   comprehensive scaffold…`). Leave the English word "scaffold" in the rest
   of the description unchanged — the developer can edit the prose later.

---

## Step 7 — Update default configuration

### `$ARGUMENTS/assets/config.default.json`

Change the `app.name` value:
```json
"name": "scaffold"
```
→
```json
"name": "$ARGUMENTS"
```

### `$ARGUMENTS/config/defaults.go`

Change the two App defaults:
```go
Name:    "template",
```
→
```go
Name:    "$ARGUMENTS",
```

```go
Title:   "Template V2 Enhanced",
```
→
```go
Title:   "$ARGUMENTS",
```

---

## Step 8 — Update `README.md`

In `$ARGUMENTS/README.md`:

1. Change the H1 title:
   ```
   # scaffold
   ```
   →
   ```
   # $ARGUMENTS
   ```

2. Replace every occurrence of the inline code `` `scaffold` `` with
   `` `$ARGUMENTS` `` (backtick-wrapped CLI invocation references only —
   do not change prose descriptions of the scaffold project itself).

---

## Step 9 — Run `go mod tidy`

Inside `$ARGUMENTS/`, run:
```
go mod tidy
```

If it fails, show the error and stop.

---

## Step 10 — Validate with `go build`

Inside `$ARGUMENTS/`, run:
```
go build ./...
```

If it fails, show the compiler output and stop.

---

## Step 11 — Report

Tell the user:

- ✅ New app created at `$ARGUMENTS/`
- Go module: `$ARGUMENTS`
- Run it: `cd $ARGUMENTS && go run .`
- The developer guide is at `scaffold/README.md`
- Suggested next steps:
  - Edit `$ARGUMENTS/cmd/root.go` — update the `Short` and `Long` descriptions
  - Edit `$ARGUMENTS/config/defaults.go` — set the display `Title`
  - Start adding screens in `$ARGUMENTS/internal/ui/screens/`
