---
title: Installation
description: Install Ninjops on your system using Homebrew, binary download, Docker, or from source
---

# Installation

Choose your preferred installation method:

## Homebrew (Recommended for macOS/Linux)

```bash
brew tap v1truv1us/tap
brew install v1truv1us/tap/ninjops
```

Notes:
- Use the fully-qualified formula name to avoid conflicts if another tap also ships `ninjops`
- Upgrade with: `brew upgrade v1truv1us/tap/ninjops`

## Download Binary

Download the latest binary for your platform from the [Releases page](https://github.com/v1truv1us/ninjops/releases).

Available platforms:
- macOS (Intel and Apple Silicon)
- Linux (amd64, arm64)
- Windows

After downloading, make the binary executable and move it to your PATH:

```bash
chmod +x ninjops
sudo mv ninjops /usr/local/bin/
```

## Docker

Pull and run the official Docker image:

```bash
docker pull ghcr.io/ninjops/ninjops:latest
docker run -p 8080:8080 ghcr.io/ninjops/ninjops:latest
```

## From Source

Build from source if you want the latest development version or need to customize the build.

### Prerequisites

- Go 1.21 or later
- Make

### Build Steps

```bash
git clone https://github.com/v1truv1us/ninjops.git
cd ninjops
make install
```

This will:
1. Download dependencies
2. Build the binary
3. Run tests
4. Install `ninjops` to your `$GOPATH/bin`

## Verify Installation

After installation, verify that Ninjops is working:

```bash
ninjops version
```

You should see the version information printed to the console.

## Next Steps

Once installed, continue to the [Quick Start guide](/getting-started/quick-start/) to create your first quote.
