package shell

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestZshInitSyntax(t *testing.T) {
	t.Parallel()

	command := exec.Command("zsh", "-n")
	command.Stdin = strings.NewReader(ZshInit)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("zsh -n failed: %v\n%s", err, output)
	}
}

func TestZshBootstrapBypassesBuiltinR(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	fakeR := filepath.Join(dir, "r")
	contents := `#!/bin/sh
if [ "$1 $2" != "init zsh" ]; then
  exit 2
fi
printf '%s\n' "typeset -g _R_BOOTSTRAP_OK=yes"
`
	if err := os.WriteFile(fakeR, []byte(contents), 0o755); err != nil {
		t.Fatalf("write fake r: %v", err)
	}

	command := exec.Command("zsh", "-dfc", `
[[ "$(whence -w r)" == 'r: builtin' ]]
eval "$(command r init zsh)"
[[ "$_R_BOOTSTRAP_OK" == 'yes' ]]
`)
	command.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("zsh bootstrap failed: %v\n%s", err, output)
	}
}

func TestZshInitCapturesCommandAndIgnoresR(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	fakeR := filepath.Join(dir, "r")
	if err := os.WriteFile(fakeR, []byte("#!/bin/sh\nprintf '%s|%s\\n' \"$R_LAST_COMMAND\" \"$*\"\n"), 0o755); err != nil {
		t.Fatalf("write fake r: %v", err)
	}

	script := `
eval "$R_INIT"
_r_capture_preexec 'ls "main file.py"'
_r_capture_preexec 'r vim'
r vim
`
	command := exec.Command("zsh", "-dfc", script)
	command.Env = append(os.Environ(), "R_INIT="+ZshInit, "PATH="+dir+":"+os.Getenv("PATH"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("zsh integration failed: %v\n%s", err, output)
	}

	want := "ls \"main file.py\"|vim\n"
	if string(output) != want {
		t.Fatalf("zsh integration output = %q, want %q", output, want)
	}
}

func TestZshInitIsSessionLocal(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	fakeR := filepath.Join(dir, "r")
	if err := os.WriteFile(fakeR, []byte("#!/bin/sh\nprintf '<%s>' \"$R_LAST_COMMAND\"\n"), 0o755); err != nil {
		t.Fatalf("write fake r: %v", err)
	}

	command := exec.Command("zsh", "-dfc", `eval "$R_INIT"; r vim`)
	command.Env = append(os.Environ(), "R_INIT="+ZshInit, "PATH="+dir+":"+os.Getenv("PATH"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("zsh integration failed: %v\n%s", err, output)
	}
	if string(output) != "<>" {
		t.Fatalf("zsh integration output = %q, want %q", output, "<>")
	}
}
