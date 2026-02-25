#!/bin/bash
#
# new-app.sh - Create a new BubbleTea v2 application from the scaffold
#
# Usage: ./scripts/new-app.sh <app-name>
# Example: ./scripts/new-app.sh myapp
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored message
info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Get the repository root (where this script is located)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# -----------------------------------------------------------------------------
# Step 1 — Validate the app name
# -----------------------------------------------------------------------------

APP_NAME="${1:-}"

if [ -z "$APP_NAME" ]; then
    error "No app name provided"
    echo ""
    echo "Usage: $0 <app-name>"
    echo "Example: $0 myapp"
    exit 1
fi

info "Validating app name: $APP_NAME"

if [[ ! "$APP_NAME" =~ ^[a-z][a-z0-9-]*$ ]]; then
    error "Invalid app name: '$APP_NAME'"
    echo ""
    echo "The name must be a valid Go identifier component:"
    echo "  - Lowercase letters, digits, and hyphens only"
    echo "  - Must start with a letter"
    echo "  - No spaces or special characters"
    exit 1
fi

# -----------------------------------------------------------------------------
# Step 2 — Check the destination
# -----------------------------------------------------------------------------

SOURCE_DIR="$REPO_ROOT/scaffold"
DEST_DIR="$REPO_ROOT/$APP_NAME"

info "Source: $SOURCE_DIR"
info "Destination: $DEST_DIR"

if [ ! -d "$SOURCE_DIR" ]; then
    error "Scaffold directory not found at: $SOURCE_DIR"
    exit 1
fi

if [ -d "$DEST_DIR" ]; then
    error "Destination already exists: $DEST_DIR"
    echo ""
    echo "Please remove the existing directory or choose a different name."
    exit 1
fi

# -----------------------------------------------------------------------------
# Step 3 — Copy the scaffold
# -----------------------------------------------------------------------------

info "Copying scaffold to $APP_NAME..."

cp -r "$SOURCE_DIR" "$DEST_DIR"

info "Copy complete"

# -----------------------------------------------------------------------------
# Step 4 — Rewrite go.mod
# -----------------------------------------------------------------------------

info "Updating go.mod..."

sed -i "s|^module scaffold$|module $APP_NAME|" "$DEST_DIR/go.mod"

# -----------------------------------------------------------------------------
# Step 5 — Update internal import paths in all Go files
# -----------------------------------------------------------------------------

info "Updating import paths in Go files..."

# Find all .go files and update import paths
find "$DEST_DIR" -name "*.go" -type f | while read -r file; do
    # Replace "scaffold/ with "$APP_NAME/
    sed -i "s|\"scaffold/|\"$APP_NAME/|g" "$file"
done

# -----------------------------------------------------------------------------
# Step 6 — Update cmd/root.go
# -----------------------------------------------------------------------------

info "Updating cmd/root.go..."

ROOT_GO="$DEST_DIR/cmd/root.go"

# 1. Update the Use field
sed -i 's|Use:   "scaffold",|Use:   "'"$APP_NAME"'",|' "$ROOT_GO"

# 2. Update example invocations (lines starting with "  scaffold")
sed -i 's|^  scaffold|  '"$APP_NAME"'|' "$ROOT_GO"

# 3. Update config file default hint
sed -i 's|\$HOME/\.scaffold\.json|$HOME/.'"$APP_NAME"'.json|' "$ROOT_GO"

# 4. Update Long description opening (first occurrence of scaffold at start of Long string)
# This is tricky - we need to find the Long: field and change only the first "scaffold" word
# The Long field typically starts with backticks and "scaffold is a..."
sed -i 's|`scaffold is a|`'"$APP_NAME"' is a|' "$ROOT_GO"

# -----------------------------------------------------------------------------
# Step 7 — Update default configuration
# -----------------------------------------------------------------------------

info "Updating default configuration..."

# Update assets/config.default.json
CONFIG_JSON="$DEST_DIR/assets/config.default.json"
if [ -f "$CONFIG_JSON" ]; then
    sed -i 's|"name": "scaffold"|"name": "'"$APP_NAME"'"|' "$CONFIG_JSON"
fi

# Update config/defaults.go
DEFAULTS_GO="$DEST_DIR/config/defaults.go"
if [ -f "$DEFAULTS_GO" ]; then
    # Update Name field
    sed -i 's|Name:    "template",|Name:    "'"$APP_NAME"'",|' "$DEFAULTS_GO"
    # Update Title field
    sed -i 's|Title:   "Template V2 Enhanced",|Title:   "'"$APP_NAME"'",|' "$DEFAULTS_GO"
fi

# -----------------------------------------------------------------------------
# Step 8 — Update README.md
# -----------------------------------------------------------------------------

info "Updating README.md..."

README="$DEST_DIR/README.md"

# 1. Change H1 title
sed -i 's|^# scaffold$|# '"$APP_NAME"'|' "$README"

# 2. Replace inline code `scaffold` with `$APP_NAME`
sed -i 's|`scaffold`|`'"$APP_NAME"'`|g' "$README"

# -----------------------------------------------------------------------------
# Step 9 — Run go mod tidy
# -----------------------------------------------------------------------------

info "Running go mod tidy..."

cd "$DEST_DIR"

if ! go mod tidy 2>&1; then
    error "go mod tidy failed"
    cd "$REPO_ROOT"
    exit 1
fi

cd "$REPO_ROOT"

# -----------------------------------------------------------------------------
# Step 10 — Validate with go build
# -----------------------------------------------------------------------------

info "Validating with go build..."

cd "$DEST_DIR"

if ! go build ./... 2>&1; then
    error "go build failed"
    cd "$REPO_ROOT"
    exit 1
fi

cd "$REPO_ROOT"

# -----------------------------------------------------------------------------
# Step 11 — Report
# -----------------------------------------------------------------------------

echo ""
echo "=========================================="
echo -e "${GREEN}New app created successfully!${NC}"
echo "=========================================="
echo ""
echo "  Location:     $APP_NAME/"
echo "  Go module:    $APP_NAME"
echo "  Run it:       cd $APP_NAME && go run ."
echo ""
echo "  Developer guide: scaffold/README.md"
echo ""
echo "Suggested next steps:"
echo "  1. Edit $APP_NAME/cmd/root.go"
echo "     - Update the Short and Long descriptions"
echo ""
echo "  2. Edit $APP_NAME/config/defaults.go"
echo "     - Set the display Title"
echo ""
echo "  3. Start adding screens in $APP_NAME/internal/ui/screens/"
echo ""