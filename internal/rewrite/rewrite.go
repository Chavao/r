package rewrite

import (
	"errors"
	"fmt"
	"strings"

	"mvdan.cc/sh/v3/shell"
	"mvdan.cc/sh/v3/syntax"
)

// ErrUnsupportedCommand reports valid shell syntax outside r's simple-command
// contract.
var ErrUnsupportedCommand = errors.New("unsupported command")

// RemainingArguments parses a command and returns every argument after its
// executable. It accepts only static simple commands so execution never needs
// to evaluate the captured command through a shell.
func RemainingArguments(command string) ([]string, error) {
	file, err := syntax.NewParser(syntax.Variant(syntax.LangZsh)).Parse(strings.NewReader(command), "")
	if err != nil {
		return nil, fmt.Errorf("parse command: %w", err)
	}
	if len(file.Stmts) != 1 {
		return nil, ErrUnsupportedCommand
	}

	stmt := file.Stmts[0]
	if stmt.Negated || stmt.Background || stmt.Coprocess || stmt.Disown || len(stmt.Redirs) != 0 {
		return nil, ErrUnsupportedCommand
	}
	call, ok := stmt.Cmd.(*syntax.CallExpr)
	if !ok || len(call.Assigns) != 0 || len(call.Args) == 0 {
		return nil, ErrUnsupportedCommand
	}
	for _, word := range call.Args {
		if !staticWord(word) {
			return nil, ErrUnsupportedCommand
		}
	}

	fields, err := shell.Fields(command, func(string) string { return "" })
	if err != nil {
		return nil, fmt.Errorf("decode command: %w", err)
	}
	if len(fields) == 0 {
		return nil, ErrUnsupportedCommand
	}
	return fields[1:], nil
}

func staticWord(word *syntax.Word) bool {
	for i, part := range word.Parts {
		switch part := part.(type) {
		case *syntax.Lit:
			if containsUnquotedExpansion(part.Value) || i == 0 && strings.HasPrefix(part.Value, "~") {
				return false
			}
		case *syntax.SglQuoted:
			if part.Dollar {
				return false
			}
		case *syntax.DblQuoted:
			if part.Dollar || !staticQuotedParts(part.Parts) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func staticQuotedParts(parts []syntax.WordPart) bool {
	for _, part := range parts {
		if _, ok := part.(*syntax.Lit); !ok {
			return false
		}
	}
	return true
}

func containsUnquotedExpansion(value string) bool {
	escaped := false
	for _, char := range value {
		if escaped {
			escaped = false
			continue
		}
		if char == '\\' {
			escaped = true
			continue
		}
		if strings.ContainsRune("*?[{}", char) {
			return true
		}
	}
	return false
}
