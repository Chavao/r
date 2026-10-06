# r

`r` is a small Linux command-line tool for rerunning the arguments from your
previous zsh command with a different executable.

It is useful when one command helps you discover, inspect, or verify a file and
your next step is to open that same file with another program.

```console
$ ls main.py
main.py
$ r vim
```

`r` turns the second command into the equivalent of:

```diff
- ls main.py
+ vim main.py
```

In the colored transformation examples throughout this document, the red line
is the captured command and the green line is the command that `r` executes.

## Why use r?

Without `r`, switching tools often means repeating or copying the same path:

```console
$ file "Quarterly Report.pdf"
Quarterly Report.pdf: PDF document, version 1.7
$ xdg-open "Quarterly Report.pdf"
```

With `r`, only the new executable is required:

```console
$ file "Quarterly Report.pdf"
Quarterly Report.pdf: PDF document, version 1.7
$ r xdg-open
```

```diff
- file "Quarterly Report.pdf"
+ xdg-open "Quarterly Report.pdf"
```

The original executable is discarded. Every argument after it is preserved,
including quoted paths and empty quoted arguments.

## Examples

### Open a discovered source file

```console
$ ls cmd/server/main.go
cmd/server/main.go
$ r nvim
```

```diff
- ls cmd/server/main.go
+ nvim cmd/server/main.go
```

### Open several files at once

```console
$ ls main.go config.go handlers.go
config.go  handlers.go  main.go
$ r code
```

```diff
- ls main.go config.go handlers.go
+ code main.go config.go handlers.go
```

### Move from a quick log preview to a pager

```console
$ head application.log
...
$ r less
```

```diff
- head application.log
+ less application.log
```

### Inspect and then open a document

```console
$ file "Architecture Notes.md"
Architecture Notes.md: Unicode text, UTF-8 text
$ r glow
```

```diff
- file "Architecture Notes.md"
+ glow "Architecture Notes.md"
```

### Open an image in a desktop application

```console
$ stat assets/system-diagram.svg
...
$ r inkscape
```

```diff
- stat assets/system-diagram.svg
+ inkscape assets/system-diagram.svg
```

### Preserve escaped spaces

```console
$ ls release\ notes.txt
release notes.txt
$ r nvim
```

```diff
- ls release\ notes.txt
+ nvim "release notes.txt"
```

## Installation

### Requirements

- Linux
- zsh
- Go 1.27.1 or newer

The initial release supports zsh only. Bash and Fish integrations are not yet
included.

### Install from the repository

```bash
git clone https://github.com/Chavao/r.git
cd r
go install ./cmd/r
```

For local development, a standalone binary can be built instead:

```bash
go build -o bin/r ./cmd/r
```

### Install with go install

Once a version has been published, it can be installed directly:

```bash
go install github.com/Chavao/r/cmd/r@latest
```

By default, Go places the binary in `$HOME/go/bin`. If `GOBIN` or `GOPATH` is
customized, use `go env GOBIN` and `go env GOPATH` to find the installation
directory.

### Configure zsh

zsh already has a built-in command named `r`, so the external binary must be
available on `PATH` before the integration is initialized. Add these lines to
`~/.zshrc` in this order:

```zsh
export PATH="$HOME/go/bin:$PATH"
eval "$(command r init zsh)"
```

If Go installs binaries elsewhere, replace `$HOME/go/bin` with that directory.
For example:

```zsh
export PATH="/custom/go/bin:$PATH"
eval "$(command r init zsh)"
```

The `command` prefix is intentional. It bypasses zsh's built-in history command
and selects the external `r` executable from `PATH` during initialization.

Place the integration after any framework or dotfiles loader that constructs
your final `PATH`, or add the Go binary directory explicitly before it. Then
start a new shell:

```zsh
exec zsh
```

Verify the setup:

```console
$ whence -va r
r is a shell function from zsh
r is a shell builtin
r is /home/you/go/bin/r
$ r --help
usage: r <executable>
       r init zsh
zsh setup: eval "$(command r init zsh)"
```

## Usage

Run an ordinary command first, then pass one replacement executable to `r`:

```console
$ <original-command> <arguments...>
$ r <replacement-executable>
```

Available commands:

| Command | Description |
|---|---|
| `r <executable>` | Replace the previous command's executable and run the result. |
| `r init zsh` | Print the zsh hook and wrapper used during shell initialization. |
| `r --help` | Print CLI usage and setup information. |

The replacement must be exactly one external executable name or path. Prefix
arguments such as `r nvim -R` are intentionally unsupported in the current
version.

### Supported previous commands

`r` accepts one static, simple command containing:

- ordinary arguments;
- single-quoted strings;
- double-quoted literal strings;
- escaped spaces and other escaped literal characters;
- multiple file arguments;
- empty quoted arguments.

Examples of accepted commands:

```zsh
ls main.go
file "Quarterly Report.pdf"
ls release\ notes.txt
printf ''
```

### Unsupported previous commands

For predictable and safe execution, `r` rejects shell syntax that would require
evaluating the captured command again:

```zsh
cat file.txt | grep text       # pipeline
cat file.txt > output.txt      # redirection
cat one && cat two             # compound command
FOO=bar cat file.txt           # inline assignment
cat *.txt                      # glob expansion
cat "$HOME/file.txt"          # parameter expansion
cat "$(find-file)"             # command substitution
cat <(generate-file)           # process substitution
```

An unsupported command is not partially executed. `r` prints an error and exits
without launching the replacement.

### Argument compatibility

`r` preserves arguments; it does not translate options between programs. For
example:

```diff
- ls -la src
+ vim -la src
```

The result may be invalid because `-la` has different meaning to `vim`. Use `r`
when the original arguments are meaningful to the replacement, most commonly
when they are file paths.

## How it works

The zsh integration and Go binary have separate responsibilities:

```text
zsh preexec hook
      |
      | captures the last eligible command in this shell session
      v
shell-local _R_LAST_COMMAND
      |
      | r wrapper exposes it only to the r child process
      v
Go CLI: parse -> validate -> replace executable -> exec
      |
      v
replacement process inherits the terminal, environment, and working directory
```

The zsh hook ignores calls to `r`, so consecutive replacements continue to use
the same original command until another eligible command runs.

The Go binary:

1. reads the captured command from `R_LAST_COMMAND`;
2. parses it using the zsh grammar from `mvdan.cc/sh/v3`;
3. rejects syntax outside the supported static-command subset;
4. resolves the replacement through `PATH`;
5. replaces itself with the target process using `syscall.Exec`.

Because the Go process is replaced, interactive applications receive the same
terminal, signals, working directory, and environment as a directly executed
command.

### Session behavior

- Command state is local to the current zsh session.
- No command-history or state file is created.
- Different terminal windows do not overwrite one another's captured command.
- The last attempted eligible command is captured even if it exits unsuccessfully.
- Running `r` does not replace the captured command.

## Exit codes

| Code | Meaning |
|---:|---|
| `0` | Help or shell initialization completed successfully. Successful replacements return the target program's status because `r` becomes that process. |
| `1` | No command was captured, parsing failed, or the previous syntax is unsupported. |
| `2` | Invalid CLI arguments. |
| `126` | The replacement was found but could not be executed. |
| `127` | The replacement executable was not found. |

## Troubleshooting

### `fc: event not found: init`

zsh's built-in `r` handled the initialization command instead of the installed
binary. Use the required `command` prefix:

```zsh
eval "$(command r init zsh)"
```

### `command not found: r` while loading `.zshrc`

The Go binary directory is added to `PATH` after the initialization line. Move
the initialization below your `PATH` setup, or add the directory first:

```zsh
export PATH="$HOME/go/bin:$PATH"
eval "$(command r init zsh)"
```

Confirm the binary location with:

```zsh
go env GOBIN
go env GOPATH
whereis r
```

### `no previous command captured`

The hook is not loaded in the current shell, or no eligible command has run
since it was loaded. Start a new shell after updating `.zshrc`, then run an
ordinary command before using `r`.

### `the previous command uses syntax unsupported by this version`

The captured command contains a pipeline, redirection, expansion, assignment,
or another unsupported construct. Run a simple command containing static
arguments and try again.

## Security and privacy

`r` does not persist captured commands to disk and does not evaluate them with
`zsh -c`. The shell-local value is exposed to the `r` child process only when
the wrapper is invoked.

As with normal shell history, avoid placing passwords, tokens, or other secrets
directly on a command line.

## Contributing

Contributions should remain focused on the command-replacement workflow and
preserve the current safety properties.

### Development setup

```bash
git clone https://github.com/Chavao/r.git
cd r
go mod download
go test ./...
```

Build a development binary:

```bash
go build -o bin/r ./cmd/r
```

To test it in a temporary zsh session, put the development binary first on
`PATH` before initializing the integration:

```zsh
export PATH="$PWD/bin:$PATH"
eval "$(command r init zsh)"
```

### Required checks

Run all checks before opening a pull request:

```bash
go fmt ./...
go vet ./...
go test ./...
go test -race ./...
go build -o bin/r ./cmd/r
```

Behavior changes should include tests. In particular:

- parser changes need accepted and rejected command cases;
- CLI changes need exit-code and output assertions;
- zsh changes need syntax validation and integration coverage;
- bug fixes should include a regression test reproducing the original failure.

### Pull request guidelines

1. Keep the change small and limited to one coherent behavior.
2. Explain the user-visible problem and the chosen solution.
3. Preserve backward compatibility unless the break is explicitly justified.
4. Do not weaken command validation or introduce shell re-evaluation silently.
5. Update this README when installation, usage, or supported syntax changes.
6. Confirm that formatting, vet, tests, race tests, and the build all pass.

## License

This project is available under the [MIT License](LICENSE).
