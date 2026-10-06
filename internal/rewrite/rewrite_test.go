package rewrite

import (
	"errors"
	"reflect"
	"testing"
)

func TestRemainingArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command string
		want    []string
	}{
		{name: "ordinary arguments", command: "ls main.py", want: []string{"main.py"}},
		{name: "double quoted path", command: `cat "my file.txt"`, want: []string{"my file.txt"}},
		{name: "single quoted path", command: `cat 'my file.txt'`, want: []string{"my file.txt"}},
		{name: "escaped space", command: `cat my\ file.txt`, want: []string{"my file.txt"}},
		{name: "empty quoted argument", command: `printf ''`, want: []string{""}},
		{name: "concatenated literals", command: `cat prefix" middle"suffix`, want: []string{"prefix middlesuffix"}},
		{name: "command without arguments", command: "ls", want: []string{}},
		{name: "quoted glob is literal", command: `cat "*.txt"`, want: []string{"*.txt"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := RemainingArguments(tt.command)
			if err != nil {
				t.Fatalf("RemainingArguments() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("RemainingArguments() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestRemainingArgumentsRejectsUnsupportedCommands(t *testing.T) {
	t.Parallel()

	tests := []string{
		"FOO=bar cat file.txt",
		"cat file.txt | grep text",
		"cat file.txt > output.txt",
		"cat one; cat two",
		"cat one && cat two",
		"cat file.txt &",
		"cat $HOME/file.txt",
		"cat $((1 + 1))",
		"cat $(printf file.txt)",
		"cat <(printf file.txt)",
		"cat *.txt",
		"cat file?.txt",
		"cat file[12].txt",
		"cat file{1,2}.txt",
		"cat ~/file.txt",
		"cat $'file.txt'",
	}

	for _, command := range tests {
		command := command
		t.Run(command, func(t *testing.T) {
			t.Parallel()

			_, err := RemainingArguments(command)
			if !errors.Is(err, ErrUnsupportedCommand) {
				t.Fatalf("RemainingArguments() error = %v, want ErrUnsupportedCommand", err)
			}
		})
	}
}

func TestRemainingArgumentsRejectsMalformedSyntax(t *testing.T) {
	t.Parallel()

	_, err := RemainingArguments(`cat "unterminated`)
	if err == nil || errors.Is(err, ErrUnsupportedCommand) {
		t.Fatalf("RemainingArguments() error = %v, want parse error", err)
	}
}
