package app

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "usage: r <executable>") {
		t.Fatalf("Run() stdout = %q, want usage", stdout.String())
	}
	if !strings.Contains(stdout.String(), `eval "$(command r init zsh)"`) {
		t.Fatalf("Run() stdout = %q, want zsh bootstrap instruction", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run() stderr = %q, want empty", stderr.String())
	}
}

func TestRunInitZsh(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	code := Run([]string{"init", "zsh"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "add-zsh-hook preexec") {
		t.Fatalf("Run() stdout does not contain zsh integration")
	}
	if stderr.Len() != 0 {
		t.Fatalf("Run() stderr = %q, want empty", stderr.String())
	}
}

func TestRunRejectsInvalidArguments(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	code := Run(nil, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("Run() code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "usage: r <executable>") {
		t.Fatalf("Run() stderr = %q, want usage", stderr.String())
	}
}

func TestRunBuildsReplacementInvocation(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotArgv, gotEnvironment []string
	sys := system{
		getenv: func(name string) string {
			if name != lastCommandEnv {
				t.Fatalf("getenv() name = %q, want %q", name, lastCommandEnv)
			}
			return `cat "my file.txt"`
		},
		environ: func() []string { return []string{"TEST=value"} },
		lookPath: func(name string) (string, error) {
			if name != "vim" {
				t.Fatalf("lookPath() name = %q, want vim", name)
			}
			return "/usr/bin/vim", nil
		},
		exec: func(path string, argv, environment []string) error {
			gotPath = path
			gotArgv = argv
			gotEnvironment = environment
			return nil
		},
	}

	var stdout, stderr bytes.Buffer
	code := run([]string{"vim"}, &stdout, &stderr, sys)
	if code != 0 {
		t.Fatalf("run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if gotPath != "/usr/bin/vim" {
		t.Fatalf("exec() path = %q, want /usr/bin/vim", gotPath)
	}
	if want := []string{"vim", "my file.txt"}; !reflect.DeepEqual(gotArgv, want) {
		t.Fatalf("exec() argv = %#v, want %#v", gotArgv, want)
	}
	if want := []string{"TEST=value"}; !reflect.DeepEqual(gotEnvironment, want) {
		t.Fatalf("exec() environment = %#v, want %#v", gotEnvironment, want)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("run() output = stdout %q, stderr %q; want empty", stdout.String(), stderr.String())
	}
}

func TestRunErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		previous   string
		lookupErr  error
		execErr    error
		wantCode   int
		wantStderr string
	}{
		{
			name:       "missing captured command",
			wantCode:   1,
			wantStderr: `eval "$(command r init zsh)"`,
		},
		{
			name:       "unsupported captured command",
			previous:   "cat file.txt | grep text",
			wantCode:   1,
			wantStderr: "syntax unsupported",
		},
		{
			name:       "replacement not found",
			previous:   "cat file.txt",
			lookupErr:  errors.New("not found"),
			wantCode:   127,
			wantStderr: "executable not found",
		},
		{
			name:       "replacement cannot execute",
			previous:   "cat file.txt",
			execErr:    errors.New("permission denied"),
			wantCode:   126,
			wantStderr: "could not execute",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sys := system{
				getenv:  func(string) string { return tt.previous },
				environ: func() []string { return nil },
				lookPath: func(string) (string, error) {
					return "/usr/bin/vim", tt.lookupErr
				},
				exec: func(string, []string, []string) error { return tt.execErr },
			}

			var stdout, stderr bytes.Buffer
			code := run([]string{"vim"}, &stdout, &stderr, sys)
			if code != tt.wantCode {
				t.Fatalf("run() code = %d, want %d", code, tt.wantCode)
			}
			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Fatalf("run() stderr = %q, want substring %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}
