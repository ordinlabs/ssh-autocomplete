# ssh-autocomplete

A fast SSH host autocompletion helper that parses your `~/.ssh/config` (including `Include` directives) and provides tab-completion for `ssh`, `scp`, `sftp`, and `ssh-copy-id` across bash, zsh, and PowerShell.

## Features

- Parses `~/.ssh/config` including nested `Include` directives and glob patterns
- Filters out wildcard entries (`*`, `?`)
- Deduplicates and sorts host names
- Caches results for 10 seconds for fast repeated completions
- Cross-platform: Linux, macOS, Windows
- Supports bash, zsh, and PowerShell
- PowerShell completions support `user@host` matching
- Automatic shell detection and profile setup
- Managed marker comments for safe updates and uninstalls

## Installation

### Homebrew (macOS/Linux)

```sh
brew install ordinlabs/tap/ssh-autocomplete
```

### Winget (Windows)

```sh
winget install OrdinLabs.ssh-autocomplete
```

### From releases

Download the latest binary from [GitHub Releases](https://github.com/ordinlabs/ssh-autocomplete/releases) and place it somewhere on your `PATH`.

### From source

```sh
go install github.com/ordinlabs/ssh-autocomplete@latest
```

## Usage

### Automatic setup

The easiest way to get started — detects your shell, writes the completion script, and adds it to your profile:

```sh
ssh-autocomplete setup
```

This will:
1. Detect your OS and shell (bash, zsh, or PowerShell)
2. Write the appropriate completion script to `~/.ordin/`
3. Ask permission before adding a source line to your shell profile
4. Wrap the addition in `# BEGIN ssh-autocomplete` / `# END ssh-autocomplete` markers for safe future updates

Running `setup` again will detect the existing block and offer to update it in place. It also detects legacy installs from older versions (without markers) and offers to migrate them.

### Manual setup

If you prefer to configure things yourself:

```sh
ssh-autocomplete setup help
```

This prints the instructions for each shell without modifying any files.

### Uninstall

Removes the completion block from your shell profile and deletes the script files:

```sh
ssh-autocomplete uninstall
```

### Generate host names

Outputs all non-wildcard host names from your SSH config, one per line:

```sh
ssh-autocomplete generate
```

Use `--no-cache` to bypass the 10-second cache:

```sh
ssh-autocomplete generate --no-cache
```

## How it works

The `generate` command parses `~/.ssh/config`, follows `Include` directives recursively, extracts all `Host` entries, filters out wildcards, and outputs the deduplicated sorted list. Completion scripts call this command when tab is pressed.

Results are cached in a temp file for 10 seconds so repeated tab presses don't re-parse the config each time.

## Development

```sh
# Run directly
go run . generate

# Setup in dev mode (uses "go run <project_dir>" in completion scripts)
go run . setup

# Build
go build -o ssh-autocomplete .
```

When run via `go run`, the setup command automatically detects dev mode and writes completion scripts that invoke `go run <project_dir> generate` instead of assuming `ssh-autocomplete` is on your PATH.

## License

MIT
