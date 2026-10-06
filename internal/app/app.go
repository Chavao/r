package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/Chavao/r/internal/rewrite"
	shellinit "github.com/Chavao/r/internal/shell"
)

const lastCommandEnv = "R_LAST_COMMAND"

type system struct {
	getenv   func(string) string
	environ  func() []string
	lookPath func(string) (string, error)
	exec     func(string, []string, []string) error
}

// Run executes the r command and returns its process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	return run(args, stdout, stderr, system{
		getenv:   os.Getenv,
		environ:  os.Environ,
		lookPath: exec.LookPath,
		exec:     syscall.Exec,
	})
}

func run(args []string, stdout, stderr io.Writer, sys system) int {
	switch {
	case len(args) == 1 && (args[0] == "-h" || args[0] == "--help"):
		printUsage(stdout)
		return 0
	case len(args) == 2 && args[0] == "init" && args[1] == "zsh":
		_, _ = io.WriteString(stdout, shellinit.ZshInit)
		return 0
	case len(args) != 1:
		printUsage(stderr)
		return 2
	}

	previous := sys.getenv(lastCommandEnv)
	if previous == "" {
		fmt.Fprintln(stderr, `r: no previous command captured; initialize zsh with: eval "$(command r init zsh)"`)
		return 1
	}

	remaining, err := rewrite.RemainingArguments(previous)
	if err != nil {
		if errors.Is(err, rewrite.ErrUnsupportedCommand) {
			fmt.Fprintln(stderr, "r: the previous command uses syntax unsupported by this version")
		} else {
			fmt.Fprintln(stderr, "r: could not parse the previous command")
		}
		return 1
	}

	replacement := args[0]
	path, err := sys.lookPath(replacement)
	if err != nil {
		fmt.Fprintf(stderr, "r: executable not found: %s\n", replacement)
		return 127
	}

	argv := make([]string, 1, len(remaining)+1)
	argv[0] = replacement
	argv = append(argv, remaining...)
	if err := sys.exec(path, argv, sys.environ()); err != nil {
		fmt.Fprintf(stderr, "r: could not execute %s: %v\n", replacement, err)
		return 126
	}

	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: r <executable>")
	fmt.Fprintln(w, "       r init zsh")
	fmt.Fprintln(w, `zsh setup: eval "$(command r init zsh)"`)
}
