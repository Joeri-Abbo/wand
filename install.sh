#!/usr/bin/env bash
# wand-install.sh: Clone, build, and set up the wand CLI tool
set -e

REPO_URL="https://github.com/Joeri-Abbo/wand"
INSTALL_DIR="$HOME/wand"
GO_VERSION="1.22.0"

# 1. Install Go if not present
if ! command -v go >/dev/null 2>&1; then
  echo "Go not found. Installing Go $GO_VERSION..."
  if [[ "$OSTYPE" == "darwin"* ]]; then
    brew install go || {
      echo "Homebrew not found. Please install Go manually."; exit 1;
    }
  else
    echo "Please install Go $GO_VERSION or later manually."
    exit 1
  fi
else
  echo "Go is already installed."
fi

# 2. Clone the repo
if [ ! -d "$INSTALL_DIR" ]; then
  git clone "$REPO_URL" "$INSTALL_DIR"
else
  echo "wand directory already exists. Pulling latest..."
  cd "$INSTALL_DIR"
  git pull
fi

cd "$INSTALL_DIR"

# 3. Build the binary
if [ -f go.mod ]; then
  go build -o wand main.go
else
  echo "go.mod not found. Please check the repo."
  exit 1
fi

# 4. Add to PATH in ~/.zshrc if not already present
if ! grep -q 'export PATH="$PATH:'"$INSTALL_DIR"'"' ~/.zshrc; then
  echo "export PATH=\"$PATH:$INSTALL_DIR\"" >> ~/.zshrc
  echo "Added $INSTALL_DIR to PATH in ~/.zshrc"
else
  echo "$INSTALL_DIR already in PATH in ~/.zshrc"
fi

echo "\nDone! Restart your terminal or run: source ~/.zshrc"
echo "You can now use the 'wand' command."
